package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	redispkg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
)

func DebugBatchPath(
	store *pg.BatchPathStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		batchID := chi.URLParam(r, "id")

		pathID, err := store.GetPathForBatch(
			r.Context(),
			batchID,
		)

		if err != nil {
			http.Error(w, "batch not found", http.StatusNotFound)
			return 
		}

		json.NewEncoder(w).Encode(map[string]string{
			"batch_id": batchID,
			"path_id" : pathID,
		})
	}
}

func DebugPathBatches(
	index *redispkg.PathIndex,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathID := chi.URLParam(r, "id")

		batches, err := index.GetBatchesForPath(
			r.Context(),
			pathID,
		)

		if err != nil {
			http.Error(w, "failed to read redis", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"path_id": pathID,
			"batches": batches,
		})
	}
}