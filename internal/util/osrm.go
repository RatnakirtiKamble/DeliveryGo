package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OSRMRouteResponse struct {
	Routes []struct {
		Duration float64 	`json:"duration"`
		Geometry any 		`json:"geometry"`
	} `json:"routes"`
}

func ComputeRoute(
	storeLat, storeLon,
	destLat, destLon float64,
	addr string,
) (int, any, error) {

	url := fmt.Sprintf(
		addr+"/route/v1/driving/%f,%f;%f,%f?geometries=geojson",
		storeLon, storeLat,
		destLon, destLat,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	var data OSRMRouteResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, nil, err 
	}

	if len(data.Routes) == 0 {
		return 0, nil, fmt.Errorf("no route found for store: (%f, %f) to dest: (%f, %f)", storeLat, storeLon, destLat, destLon)
	}

	return int(data.Routes[0].Duration), data.Routes[0].Geometry, nil
}

func SnapToRoad(
	lat, lon float64,
	addr string,
) (float64, float64, error) {
	url := fmt.Sprintf(
		addr+"/nearest/v1/driving/%f,%f",
		lon, lat,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err 
	}

	defer resp.Body.Close()

	var data struct{
		Waypoints []struct {
			Location [2]float64 `json:"location"`
		} `json:"waypoints"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil{
		return 0, 0, err 
	}

	if len(data.Waypoints) == 0 {
		return 0, 0, fmt.Errorf("no nearby road")
	}

	return data.Waypoints[0].Location[1], data.Waypoints[0].Location[0], nil

}