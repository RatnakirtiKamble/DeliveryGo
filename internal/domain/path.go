package domain

import "encoding/json"

type PathTemplate struct {
	ID       string          `json:"id"`
	StoreID  string          `json:"store_id"`
	H3Cell   string          `json:"h3_cell"`
	BaseETA  int             `json:"base_eta"`
	Polyline json.RawMessage `json:"polyline"`
}
