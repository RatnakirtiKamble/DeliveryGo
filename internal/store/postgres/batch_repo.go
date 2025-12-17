package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type BatchStore struct {
	db *pgxpool.Pool
}

func NewBatchStore(db *pgxpool.Pool) *BatchStore {
	return &BatchStore{db: db}
}

func (s *BatchStore) CreateOpenBatchWithOrder(
	ctx context.Context,
	orderID string,
) (*domain.Batch, error) {
	
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	batch := &domain.Batch{
		ID:			uuid.NewString(),
		Status: 	domain.BatchOpen,
		CreatedAt:	time.Now(),
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO batches (id, status, created_at)
		VALUES ($1, $2, $3)`,
		batch.ID,
		batch.Status,
		batch.CreatedAt,
	)	

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO batch_orders (batch_id, order_id)
		VALUES ($1, $2)`,
		batch.ID,
		orderID,
	)

	if err != nil {
		return nil, err 
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return batch, nil

}

func (s *BatchStore) GetOpenBatch(ctx context.Context) (*domain.Batch, error) {
	row := s.db.QueryRow(ctx,
	`SELECT id, status, created_at
	FROM batches
	WHERE status = 'OPEN'
	ORDER BY created_at
	LIMIT 1`)

	var batch domain.Batch
	if err := row.Scan(&batch.ID, &batch.Status, &batch.CreatedAt); err != nil {
		return nil, err
	}

	return &batch, nil
}

func (s *BatchStore) AddOrderToBatch (
	ctx context.Context,
	batchID, orderID string,
) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO batch_orders (batch_id, order_id)
		VALUES ($1, $2)`,
	batchID,
	orderID)

	return err
}