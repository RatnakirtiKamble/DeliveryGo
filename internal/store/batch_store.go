package store 

import (
	"context"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type BatchStore interface {
	GetOpenBatch(ctx context.Context) (*domain.Batch, error)
	AddOrderToBatch(ctx context.Context, batchID, orderID string) error
	CreateOpenBatchWithOrder(ctx context.Context, orderID string,) (*domain.Batch, error)
}

