package domain

import "time"

type RiderStatus string 

const (
	RiderIdle			RiderStatus = "IDLE"
	RiderAssigned   	RiderStatus = "ASSIGNED"
	RirderDelivering	RiderStatus = "DELIVERED"
)

type Rider struct {
	ID			string 
	Status 		RiderStatus
	Lat 		float64 
	Lon 		float64 
	UpdatedAt 	time.Time 
}

