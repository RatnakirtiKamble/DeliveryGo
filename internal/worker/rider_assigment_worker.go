package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
)

type RiderAssignEvent struct {
	BatchID string 	`json:"batch_id"`
	PathID  string  `json:"path_id"`
}

type RiderAssignmentWorker struct {
	riderStore *postgres.RiderStore
	riderCache *redis.RiderCache
}

func NewRiderAssignmentWorker(
	riderStore *postgres.RiderStore,
	riderCache *redis.RiderCache,
) *RiderAssignmentWorker {
	return &RiderAssignmentWorker{
		riderStore: riderStore,
		riderCache: riderCache,
	}
}

func (w *RiderAssignmentWorker) Handle(
	ctx context.Context,
	msg kafka.Message,
) error {

	if msg.Topic != "routes.refined.requested" {
		return nil
	}

	var evt RiderAssignEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err
	}

	lat, lon := 19.099822, 73.006617

	riderID, err := w.riderCache.NearestRider(ctx, lat, lon)
	if err != nil || riderID == "" {
		log.Println("[rider-assign] no available riders", err)
		return nil 
	}

	if err := w.riderStore.AssignRiderTx(
		ctx,
		evt.BatchID,
		riderID,
	); err != nil {
		return err 
	}

	_ = w.riderCache.RemoveAvailable(ctx, riderID)

	log.Printf(
		"[rider-assign] batch=%s rider=%s",
		evt.BatchID,
		riderID,
	)

	return nil
}