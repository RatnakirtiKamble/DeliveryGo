package batch

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
)

type Service struct {
	store store.BatchStore
	batchPathStore *postgres.BatchPathStore
}

func NewService(
	store store.BatchStore,
	batchPathStore *postgres.BatchPathStore,) *Service {
	return &Service{
		store: 			store,
		batchPathStore: batchPathStore,}
}

func (s *Service) AssignOrder(
	ctx context.Context,
	orderID string,
) (*domain.Batch, error) {
	return s.store.AssignOrderTx(ctx, orderID)
}

func (s *Service) AssignBatchToPath(
	ctx context.Context,
	batchID, pathID string,
) error {
	return s.batchPathStore.AssignBatchToPathOptimistic(
		ctx,
		batchID,
		pathID,
	)
}

