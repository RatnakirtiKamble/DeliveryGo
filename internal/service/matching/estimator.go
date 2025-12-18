package matching

import (
	"math"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type SimpleEstimator struct {}

func (e *SimpleEstimator) DeltaCost(
	path *domain.PathTemplate,
	order *domain.Order,
) int {
	min := math.MaxFloat64 
	for _, p := range path.Polyline {
		d := math.Hypot(p.Lat-order.Lat, p.Lon-order.Lon)
		if d < min {
			min = d
		}
	}

	return int(min * 1000)
}