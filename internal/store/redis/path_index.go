package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type PathIndex struct {
	rdb *redis.Client
}

func NewPathIndex(rdb *redis.Client) *PathIndex {
	return &PathIndex{rdb: rdb}
}

func (p *PathIndex) GetCandidatePaths(
	ctx context.Context,
	h3 string,
) ([]string, error) {
	return p.rdb.SMembers(ctx, "h3:"+h3).Result()
}

func (p *PathIndex) BindBatchToPath(
	ctx context.Context,
	pathID, batchID string,
) error {
	key := "path:" + pathID + ":batches"

	if err := p.rdb.SAdd(ctx, key, batchID).Err(); err != nil {
		return err 
	}

	return p.rdb.Expire(ctx, key, time.Hour).Err()
}

func (p *PathIndex) RemoveBatchFromPath(
	ctx context.Context,
	pathID, batchID string,
) error {
	return p.rdb.SRem(ctx, "path:"+pathID+":batches", batchID).Err()
}

func (p *PathIndex) GetBatchesForPath(
	ctx context.Context,
	pathID string,
) ([]string, error) {
	return p.rdb.SMembers(
		ctx,
		"path:"+pathID+":batches",
	).Result()
}

func (p *PathIndex) AddPathToH3(
	ctx context.Context,
	h3, pathID string,
) error {
	return p.rdb.SAdd(ctx, "h3:"+h3, pathID).Err()
}

func (p *PathIndex) RemovePathH3Cell(
	ctx context.Context,
	pathID, h3Cell string,
) error {
	return p.rdb.SRem(ctx, "h3:"+h3Cell, pathID).Err()
}