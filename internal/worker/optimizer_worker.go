package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	redispkg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	"github.com/segmentio/kafka-go"
)

type RouteRefineRequestedEvent struct {
	BatchID       string   `json:"batch_id"`
	PathID        string   `json:"path_id"`
	Orders        []string `json:"orders"`
	EstimatedCost int      `json:"estimated_cost"`
}

type OptimizerWorker struct {
	producer  *kafkaq.Producer
	pathIndex *redispkg.PathIndex
	batchPaths *postgres.BatchPathStore
}

func NewOptimizerWorker(
	producer *kafkaq.Producer,
	pathIndex *redispkg.PathIndex,
	batchPaths *postgres.BatchPathStore,
) *OptimizerWorker {
	return &OptimizerWorker{
		producer:  	producer,
		pathIndex: 	pathIndex,
		batchPaths: batchPaths,
	}
}

func (w *OptimizerWorker) Handle(
	ctx context.Context,
	msg kafka.Message,
) error {

	if msg.Topic != "routes.refine.requested" {
		return nil
	}

	var evt RouteRefineRequestedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err
	}

	log.Printf(
		"[optimizer-worker] refining route batch=%s path=%s cost=%d",
		evt.BatchID,
		evt.PathID,
		evt.EstimatedCost,
	)

	start := time.Now()

	time.Sleep(200 * time.Millisecond)
	refinedCost := evt.EstimatedCost + 15
	elapsed := time.Since(start).Milliseconds()

	if err := w.batchPaths.UpsertBatchPath(
		ctx, 
		evt.BatchID,
		evt.PathID,
	); err != nil {
		return err
	}

	if err := w.pathIndex.BindBatchToPath(
		ctx,
		evt.PathID,
		evt.BatchID,
	); err != nil {
		return err
	}

	log.Printf(
		"[optimizer-worker] bound batch=%s to path=%s in redis",
		evt.BatchID,
		evt.PathID,
	)

	out := map[string]any{
		"batch_id":         evt.BatchID,
		"path_id":          evt.PathID,
		"estimated_cost":  	evt.EstimatedCost,
		"refined_cost":    	refinedCost,
		"optimization_ms": 	elapsed,
	}

	payload, _ := json.Marshal(out)

	if err := w.producer.Publish(
		ctx,
		"routes.refined.requested",
		evt.BatchID,
		payload,
	); err != nil {
		return err
	}

	log.Printf(
		"[optimizer-worker] refinement published batch=%s refined_cost=%d",
		evt.BatchID,
		refinedCost,
	)

	return nil
}
