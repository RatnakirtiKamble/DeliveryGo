package matching

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type Client interface {
	SelectBestPath(
		ctx context.Context,
		order *domain.Order,
		candidatePathIDs []string,
	) (*domain.PathTemplate, int, error)
}
