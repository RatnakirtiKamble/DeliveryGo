package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type OrderStore struct {
	db *pgxpool.Pool 
}

func NewOrderStore(db *pgxpool.Pool) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Create(ctx context.Context, order *domain.Order) error {
	_, err := s.db.Exec(
		ctx,
		`INSERT INTO orders (id, user_id, lat, lon, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		order.ID,
		order.UserID,
		order.Lat,
		order.Lon,
		order.CreatedAt,
	)
	return err
}