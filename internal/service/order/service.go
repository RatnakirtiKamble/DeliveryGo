package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store"
)

type Service struct {
	store store.OrderStore
}

func NewService(store store.OrderStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) CreateOrder(
	ctx context.Context,
	userID string,
	lat, lon float64,
) (*domain.Order, error) {
	order := &domain.Order{
		ID:        	uuid.New().String(),
		UserID:    	userID,
		Lat:  		lat,
		Lon: 		lon,
		CreatedAt: 	time.Now(),
	}

	err := s.store.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}