package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type RouteRefinedRequestedEvent struct {
	BatchID 		string 	`json:"batch_id"`
	PathID 			string 	`json:"path_id"`
	EstimatedCost   int 	`json:"estimated_cost"`
	RefinedCost 	int 	`json:"refined_cost"`
	OptimizationMs  int64   `json:"optimization_ms"`
}

type RegretWorker struct {
	regretThreshold float64
}

func NewRegretWorker() *RegretWorker {
	return &RegretWorker{
		regretThreshold: 0.20,
	}
}

func (w *RegretWorker) Handle(
	_ context.Context,
	msg kafka.Message,
) error {

	if msg.Topic != "routes.refined.requested" {
		return nil 
	}

	var evt RouteRefinedRequestedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err 
	}

	regret := float64(evt.RefinedCost-evt.EstimatedCost) / float64(evt.EstimatedCost)

	log.Printf(
		"[regret-worker] batch=%s est=%d refined=%d regret=%.2f",
		evt.BatchID,
		evt.EstimatedCost,
		evt.RefinedCost,
		regret,
	)

	if regret > w.regretThreshold {
		log.Printf(
			"[regret-worker] regret exceeded batch=%s threshold=%.2f",
			evt.BatchID,
			w.regretThreshold,
		)
	}

	return nil
}