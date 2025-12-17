package memory

import (
	"context"
	"sync"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type OrderStore struct {
	mu         sync.Mutex
	orders     map[string]*domain.Order
}

func NewOrderStore() *OrderStore {
	return &OrderStore{
		orders: make(map[string]*domain.Order),
	}
}

func (s *OrderStore) Create(ctx context.Context, order *domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.ID] = order
	return nil
}