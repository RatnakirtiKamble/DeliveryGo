package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"log"

	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/util"
)

type confirmDeliveryReq struct {
	RiderID string  `json:"rider_id"`
	Lat     float64 `json:"lat"`
	Lon     float64  `json:"lon"`
}

func ConfirmDelivery(
	riderCache *redis.RiderCache,
	batchPathStore *pg.BatchPathStore,
	riderStore *pg.RiderStore,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		batchID := chi.URLParam(r, "id")

		var req confirmDeliveryReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}	

		pathID, err := batchPathStore.GetPathForBatch(
			r.Context(),
			batchID,
		)
		if err != nil {
			http.Error(w, "batch not found", http.StatusNotFound)
			return
		}
		
		log.Printf("[delivery-confirmation] delivered to %v", pathID)

		// [REMINDER] CHANGE THIS TO ACTUAL COORDS, Will use simulated for now
		dist := util.HaversineMeters(
			req.Lat,
			req.Lon,
			req.Lat,
			req.Lon,
		)

		if dist > 30 {
			http.Error(
				w, 
				"rider not at drop location",
				http.StatusConflict,
			)

			return
		}

		if err := riderStore.ConfirmDeliveryTX(
			r.Context(),
			batchID,
			req.RiderID,
		); err != nil {
			http.Error(w, "failed to confirm delivery", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)


	}
}
