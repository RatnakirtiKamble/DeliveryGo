package postgres 

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

)

type BatchPathStore struct {
	pool *pgxpool.Pool
}

func NewBatchPathStore(pool *pgxpool.Pool) *BatchPathStore {
	return &BatchPathStore{pool: pool}
}

func (s* BatchPathStore) UpsertBatchPath(
	ctx context.Context,
	batchID, pathID string,
) error {

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO batch_paths (batch_id, path_id)
		VALUES ($1, $2)
		ON CONFLICT (batch_id)
		DO UPDATE SET 
			path_id = EXCLUDED.path_id,
			assigned_at = now()
		`,
		batchID,
		pathID,
	)

	return err
}

func (s* BatchPathStore) GetPathForBatch(
	ctx context.Context,
	batchID string,
) (string, error) {

	var pathID string
	
	err := s.pool.QueryRow(
		ctx,
		`SELECT path_id FROM batch_paths WHERE batch_id = $1`,
		batchID,
	).Scan(&pathID)

	return pathID, err
}