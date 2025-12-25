package redis

import (
	"context"
	"strconv"
	"time"
	"log"

	"github.com/redis/go-redis/v9"
)

type RiderCache struct {
	rdb *redis.Client 
}

func NewRiderCache(rdb *redis.Client) *RiderCache {
	return &RiderCache{rdb: rdb}
}

func (c *RiderCache) UpdateLocation(
	ctx context.Context,
	riderID string, 
	lat, lon float64,
) error {

	pipe := c.rdb.TxPipeline()

	pipe.GeoAdd(ctx, "riders:available", &redis.GeoLocation{
		Name: 		riderID,
		Latitude: 	lat,
		Longitude:  lon,
	})

	pipe.HSet(ctx, "rider:"+riderID+":loc",
	"lat", strconv.FormatFloat(lat, 'f', 6, 64),
	"lon", strconv.FormatFloat(lon, 'f', 6, 64),
)

	pipe.Expire(ctx, "rider:"+riderID+":loc", time.Minute)

	_, err := pipe.Exec(ctx)

	log.Println("[redis] successfully updated cache")
	return err 
}

func (c *RiderCache) NearestRider(
	ctx context.Context,
	lat, lon float64, 
) (string, error) {

	res, err := c.rdb.GeoSearch(
		ctx,
		"riders:available",
		&redis.GeoSearchQuery{
			Longitude:  lon,
			Latitude:   lat,
			Radius:     5,
			RadiusUnit: "km",
			Count:      1,
			Sort:       "ASC",
		},
	).Result()

	if err != nil || len(res) == 0 {
		return "", err 
	}

	return res[0], nil 
}

func (c *RiderCache) RemoveAvailable(
	ctx context.Context,
	riderID string,
) error {
	return c.rdb.ZRem(ctx, "riders:available", riderID).Err()
}