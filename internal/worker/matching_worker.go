package worker 

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type BatchAssignedEvent struct {
	BatchID 		string 		`json:"batch_id"`
	PathID 			string 		`json:"path_id"`
	Orders 			[]string 	`json:"orders"`
	EstimatedCost 	int 		`json:"estimated_cost"`
}

type batchState struct {
	BatchID 		string 
	PathID 			string 		
	OrderIDs		[]string 
	EstimatedCost 	int 
	AssignedAt 		time.Time 
}

type MatchingWorker struct {
	mu 		sync.Mutex
	batches map[string]*batchState
}

func NewMatchingWorker() *MatchingWorker {
	return &MatchingWorker{
		batches: make(map[string]*batchState),
	}
}

func (w *MatchingWorker) Handle(
	_ context.Context,
	msg kafka.Message,
) error {
	
	if msg.Topic != "batches.assigned" {
		return nil 
	}

	var evt BatchAssignedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err 
	}

	log.Printf(
		"[matching-worker] batch assigned batch=%s path=%s orders=%d",
		evt.BatchID,
		evt.PathID,
		len(evt.Orders),
	)

	w.mu.Lock()
	defer w.mu.Unlock()

	w.batches[evt.BatchID] = &batchState{
		BatchID: 		evt.BatchID,
		PathID: 		evt.PathID,
		OrderIDs:  		evt.Orders,
		EstimatedCost:  evt.EstimatedCost,
		AssignedAt:  	time.Now(),
	}

	return nil
}