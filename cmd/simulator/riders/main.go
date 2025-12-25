package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"net/http"
)

const (
	apiBase = "http://localhost:8000"
	riderID = "rider-2"
)

type routeResp struct {
	BatchID string          `json:"batch_id"`
	PathID  string          `json:"path_id"`
	Polyline json.RawMessage `json:"polyline"`
}

type geoJSONLineString struct {
	Type        string        `json:"type"`
	Coordinates [][]float64   `json:"coordinates"`
}

type locPayload struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type confirmPayload struct {
	RiderID string  `json:"rider_id"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

func waitForAssignment() string {
	for {
		resp, err := http.Get(apiBase + "/debug/rider/" + riderID + "/batch")
		if err != nil || resp.StatusCode != 200 {
			time.Sleep((2 * time.Second))
			continue
		}

		var data struct {
			BatchID string `json:"batch_id"`
		}

		_ = json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		if data.BatchID != "" {
			return data.BatchID
		}

		time.Sleep(2 * time.Second)
	}
}

func fetchRoute(batchID string) geoJSONLineString {
	resp, err := http.Get(apiBase + "/debug/batch/" + batchID + "/route")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var r routeResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Fatal(err)
	}

	var line geoJSONLineString
	if err := json.Unmarshal(r.Polyline, &line); err != nil {
		log.Fatal(err)
	}

	return line
}

func followRoute(batchID string, route geoJSONLineString) {
	for _, c := range route.Coordinates {
		lat := c[1]
		lon := c[0]

		sendLocation(lat, lon)
		time.Sleep(500 * time.Millisecond)
	}

	last := route.Coordinates[len(route.Coordinates) - 1]
	confirmDelivery(batchID, last[1], last[0])
}

func sendLocation(lat, lon float64) {
	body, _ := json.Marshal(locPayload{
		Lat: lat,
		Lon: lon,
	})

	_, err := http.Post(
		apiBase+"/riders/"+riderID+"/location",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		log.Println("[sim] gps update failed:", err)
	}
}

func confirmDelivery(batchID string, lat, lon float64) {
	body, _ := json.Marshal(confirmPayload{
		RiderID: riderID,
		Lat:	 lat,
		Lon:     lon,  
	})

	resp, err := http.Post(
		apiBase+"/batches/"+batchID+"/confirm-delivery",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		log.Println("[sim] delivery confirm failed: ", err)
		return 
	}

	resp.Body.Close()

	log.Println("[sim] delivery confirmed for batch:", batchID)
}

func main() {
	log.Println("[sim] rider simulator started")


	startLon := 19.096378
	startLat := 73.005819
	sendLocation(startLat, startLon)

	log.Println("[sim] rider online, waiting for assignment")
	for {
		batchID := waitForAssignment()
		log.Println("[sim] assigned batch:", batchID)

		route := fetchRoute(batchID)
		followRoute(batchID, route)

		log.Println("[sim] delivery completed, waiting for next batch")
	}
}
