package store

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type OrderStore interface {
	Create(ctx context.Context, order *domain.Order) error
}

type BatchStore interface {
	AssignOrderTx(ctx context.Context, orderID string) (*domain.Batch, error)
}

type BatchPathStore interface {
	UpsertBatchPath(ctx context.Context, batchID, pathID string) error
	GetPathForBatch(ctx context.Context, batchID string) (string, error)
}