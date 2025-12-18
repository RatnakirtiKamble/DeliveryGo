package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/segmentio/kafka-go"
)

type RouteRefineRequestedEvent struct {
	BatchID       string   `json:"batch_id"`
	PathID        string   `json:"path_id"`
	Orders        []string `json:"orders"`
	EstimatedCost int      `json:"estimated_cost"`
}

type OptimizerWorker struct {
	producer *kafkaq.Producer
}

func NewOptimizerWorker(producer *kafkaq.Producer) *OptimizerWorker {
	return &OptimizerWorker{
		producer: producer,
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

	out := map[string]any{
		"batch_id":        evt.BatchID,
		"path_id":         evt.PathID,
		"estimated_cost": evt.EstimatedCost,
		"refined_cost":   refinedCost,
		"optimization_ms": elapsed,
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
