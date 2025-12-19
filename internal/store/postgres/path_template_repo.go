package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PathTemplateStore struct {
	pool *pgxpool.Pool
}

func NewPathTemplateStore(pool *pgxpool.Pool) *PathTemplateStore {
	return &PathTemplateStore{pool: pool}
}

func (s *PathTemplateStore) Insert(
	ctx context.Context,
	id, storeId, h3 string,
	baseETA int,
	polyline any,
) error {

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO path_templates (
			id, store_id, h3_cell, base_eta, polyline
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (store_id, h3_cell) DO NOTHING	
		`,
		id, storeId, h3, baseETA, polyline,
	)

	return err
}