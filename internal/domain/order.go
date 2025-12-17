package domain 

import "time"

type Order struct {
	ID 	  		string
	UserID 		string
	Lat    		float64
	Lon    		float64
	CreatedAt 	time.Time
}

