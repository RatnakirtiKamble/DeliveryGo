package postgres

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
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

func (s *PathTemplateStore) ListAll(
	ctx context.Context,
) (map[string]*domain.PathTemplate, error) {

	rows, err := s.pool.Query(
		ctx,
		`
		SELECT id, store_id, h3_cell, base_eta, polyline
		FROM path_templates
		`,
	)

	if err != nil{
		return nil, err
	}

	defer rows.Close()

	paths := make(map[string]*domain.PathTemplate)

	for rows.Next() {
		var p domain.PathTemplate

		if err := rows.Scan(
			&p.ID,
			&p.StoreID,
			&p.H3Cell,
			&p.BaseETA,
			&p.Polyline,
		); err != nil {
			return nil, err
		}
		paths[p.ID] = &p 
	}

	return paths, nil
}

func (s *PathTemplateStore) GetByID(
	ctx context.Context,
	id string,
) (*domain.PathTemplate, error) {

	var p domain.PathTemplate

	err := s.pool.QueryRow(
		ctx,
		`SELECT id, store_id, h3_cell, base_eta, polyline
		 FROM path_templates
		 WHERE id = $1`,
		id,
	).Scan(
		&p.ID,
		&p.StoreID,
		&p.H3Cell,
		&p.BaseETA,
		&p.Polyline,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}
