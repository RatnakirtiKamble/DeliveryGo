package matching

import (
	"encoding/json"
	"math"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type SimpleEstimator struct{}

// GeoJSON LineString shape returned by OSRM
type geoJSONLineString struct {
	Type        string        `json:"type"`
	Coordinates [][]float64   `json:"coordinates"` 
}

func (e *SimpleEstimator) DeltaCost(
	path *domain.PathTemplate,
	order *domain.Order,
) int {

	var geom geoJSONLineString
	if err := json.Unmarshal(path.Polyline, &geom); err != nil {
		return math.MaxInt
	}

	min := math.MaxFloat64

	for _, coord := range geom.Coordinates {
		if len(coord) < 2 {
			continue
		}

		lon := coord[0]
		lat := coord[1]

		d := math.Hypot(lat-order.Lat, lon-order.Lon)
		if d < min {
			min = d
		}
	}

	cost := int(min * 1000)
	if cost < 1 {
		cost = 1
	}
	return cost
}
