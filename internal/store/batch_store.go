package store

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type BatchStore interface {
	AssignOrderTx(ctx context.Context, orderID string) (*domain.Batch, error)
}
