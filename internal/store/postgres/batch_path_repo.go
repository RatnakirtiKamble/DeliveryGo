package postgres 

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

)

var (
	ErrOptimisticConflict	= errors.New("optimistic conflict")
	ErrBatchAtCapacity 		= errors.New("batch at capacity")
	ErrBatchNotOpen			= errors.New("batch not open")
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

func (s *BatchPathStore) AssignBatchToPathOptimistic(
	ctx context.Context,
	batchID string,
	pathID string,
) error {

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var (
		status      string
		capacity    int
		currentSize int
		version     int
	)

	err = tx.QueryRow(
		ctx,
		`SELECT status, capacity, current_size, version
		 FROM batches
		 WHERE id = $1`,
		batchID,
	).Scan(&status, &capacity, &currentSize, &version)
	if err != nil {
		return err
	}

	if status != "OPEN" {
		return ErrBatchNotOpen
	}

	if currentSize >= capacity {
		return ErrBatchAtCapacity
	}

	res, err := tx.Exec(
		ctx,
		`INSERT INTO batch_paths (batch_id, path_id)
		 VALUES ($1, $2)
		 ON CONFLICT (batch_id) DO NOTHING`,
		batchID,
		pathID,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrOptimisticConflict
	}

	cmd, err := tx.Exec(
		ctx,
		`UPDATE batches
		 SET current_size = current_size + 1,
		     status = 'ASSIGNED',
		     version = version + 1
		 WHERE id = $1
		   AND version = $2`,
		batchID,
		version,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrOptimisticConflict
	}

	return tx.Commit(ctx)
}
