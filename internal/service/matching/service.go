package matching

import (
	"errors"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type PathCostEstimator interface {
	DeltaCost(
		path *domain.PathTemplate,
		order *domain.Order,
	) int 
}

type Service struct {
	paths		map[string]*domain.PathTemplate
	estimator   PathCostEstimator
} 

func NewService(
	paths map[string]*domain.PathTemplate,
	estimator PathCostEstimator,
) *Service {
	return &Service{
		paths:		paths,
		estimator: 	estimator,
	}
}

func (s *Service) SelectBestPath(
	order *domain.Order,
	candidatePathIDs []string,
) (*domain.PathTemplate, int, error) {

	bestCost := int(^uint(0) >> 1)
	var bestPath *domain.PathTemplate

	for _, pid := range candidatePathIDs {
		path := s.paths[pid]
		if path == nil {
			continue
		}

		delta := s.estimator.DeltaCost(path, order)
		if delta < bestCost {
			bestCost = delta
			bestPath = path 
		}
	}

	if bestPath == nil {
		return nil, 0, errors.New("no viable path")
	}

	return bestPath, bestCost, nil

}