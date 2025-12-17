package batch

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store"
)

type Service struct {
	store store.BatchStore
}

func NewService(store store.BatchStore) *Service {
	return &Service{store: store}
}

func (s *Service) AssignOrder(
	ctx context.Context,
	orderID string,
) (*domain.Batch, error) {
	return s.store.AssignOrderTx(ctx, orderID)
}

