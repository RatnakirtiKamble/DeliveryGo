package domain

import "time"

type BatchStatus string 

const (
	BatchOpen		BatchStatus = "OPEN"
	BatchDispatched BatchStatus = "DISPATCHED"
)

type Batch struct {
	ID			string
	Status 		BatchStatus
	CreatedAt 	time.Time 
}

