package redis

import (
	"context"

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
	return p.rdb.SAdd(ctx, "path:"+pathID+":batches", batchID).Err()
}