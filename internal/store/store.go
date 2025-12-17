package store

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) error
}

