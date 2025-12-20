package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
)

type riderLocationReq struct {
	Lat	float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func UpdateRiderLocation(
	cache *redis.RiderCache,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		riderID := chi.URLParam(r, "id")

		var req riderLocationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if err := cache.UpdateLocation(
			r.Context(),
			riderID,
			req.Lat,
			req.Lon,
		); err != nil {
			http.Error(w, "failed to update location", http.StatusInternalServerError)
			return 
		}

		w.WriteHeader(http.StatusOK)
	}
}