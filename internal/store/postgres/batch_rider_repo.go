package postgres

import (
	"context"
	"time"
)

func (s *RiderStore) ConfirmDeliveryTX(
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
		`UPDATE batch_riders
		SET delivered_at = $1
		WHERE batch_id = $2 AND rider_id = $3
			AND delivered_at IS NULL`,
		time.Now(),
		batchID,
		riderID,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE riders
		SET status = 'IDLE', updated_at = $1
		WHERE id = $2`,
		time.Now(),
		riderID, 
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}