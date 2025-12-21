package main 

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type loc struct {
	Lat 	float64  `json:"lat"`
	Lon 	float64  `json:"lon"`
}

func main(){
	var riderID string = "rider-1"

	var lat float64 = 19.089
	var lon float64 = 73.002

	for {
		lat += 0.0005
		lon += 0.0003

		body, _ := json.Marshal(loc{Lat: lat, Lon: lon})

		_, err := http.Post(
			"http://localhost:8000/riders/"+riderID+"/location",
			"application/json",
			bytes.NewReader(body),
		)

		if err != nil {
			log.Println("update failed:", err)
		}

		time.Sleep(2 * time.Second)
	}
}