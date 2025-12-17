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

// func (s *BatchStore) GetOpenBatch(ctx context.Context) (*domain.Batch, error) {
// 	row := s.db.QueryRow(ctx,
// 	`SELECT id, status, created_at
// 	FROM batches
// 	WHERE status = 'OPEN'
// 	ORDER BY created_at
// 	LIMIT 1`)

// 	var batch domain.Batch
// 	if err := row.Scan(&batch.ID, &batch.Status, &batch.CreatedAt); err != nil {
// 		return nil, err
// 	}

// 	return &batch, nil
// }

// func (s *BatchStore) AddOrderToBatch (
// 	ctx context.Context,
// 	batchID, orderID string,
// ) error {
// 	_, err := s.db.Exec(ctx,
// 		`INSERT INTO batch_orders (batch_id, order_id)
// 		VALUES ($1, $2)`,
// 	batchID,
// 	orderID)

// 	return err
// }

func (s *BatchStore) AssignOrderTx(
	ctx context.Context,
	orderID string,
) (*domain.Batch, error) {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	var batch domain.Batch

	err = tx.QueryRow(ctx,
		`SELECT id, status, capacity, current_size, created_at FROM batches
		WHERE status = 'OPEN'
		AND current_size < capacity
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE
		`).Scan(
			&batch.ID,
			&batch.Status,
			&batch.Capacity,
			&batch.CurrentSize,
			&batch.CreatedAt,
		)

		if err == nil {
			_, err = tx.Exec(ctx,
			`INSERT INTO batch_orders (batch_id, order_id)
			VALUES ($1, $2)
			`,batch.ID, orderID)
			if err != nil {
				return nil, err 
			}

			_, err = tx.Exec(ctx, `
			UPDATE batches
			SET current_size = current_size + 1
			WHERE id = $1
			`, batch.ID)

			if err != nil {
				return nil, err
			}

			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}

			batch.CurrentSize++
			return &batch, nil
		}

		batch = domain.Batch{
			ID:				uuid.NewString(),
			Status: 		domain.BatchOpen,
			Capacity: 		5,
			CurrentSize: 	1,
			CreatedAt: 		time.Now(),
		}

		_, err = tx.Exec(ctx, `
		INSERT INTO Batches (id, status, capacity, current_size, created_at)
		VALUES ($1, $2, $3, $4, $5)
		`, batch.ID, batch.Status, batch.Capacity, batch.CurrentSize, batch.CreatedAt)

		if err != nil {
			return nil, err 
		}

		_, err = tx.Exec(ctx, `
		INSERT INTO batch_orders (batch_id, order_id)
		VALUES ($1, $2)
		`, batch.ID, orderID)

		if err != nil {
			return nil, err 
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err 
		}

		return &batch, nil
}