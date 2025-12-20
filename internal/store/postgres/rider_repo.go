package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RiderStore struct {
	pool *pgxpool.Pool
}

func NewRiderStore(pool *pgxpool.Pool) *RiderStore {
	return &RiderStore{pool: pool}
}

func (s *RiderStore) AssignRiderTx(
	ctx context.Context,
	batchID, riderID string,
) error {

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err 
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE riders
		SET status = 'ASSIGNED',
			updated_at = $1
		WHERE id = $2
			AND status = 'IDLE'
		`,
		time.Now(),
		riderID,
	)

	if err != nil {
		return err 
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO batch_riders (batch_id, rider_id)
		VALUES ($1, $2)`,
		batchID,
		riderID,
	)

	if err != nil {
		return err 
	}

	return tx.Commit(ctx)


}