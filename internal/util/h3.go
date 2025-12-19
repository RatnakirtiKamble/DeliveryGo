package util

import "github.com/uber/h3-go/v4"

const DefaultH3Resolution = 8

func LatLonToH3(lat, lon float64) string {
	cell, err := h3.LatLngToCell(
		h3.LatLng{
			Lat: lat,
			Lng: lon,
		},
		DefaultH3Resolution,
	)

	if err != nil{
		return ""
	}
	return cell.String()
}

func H3CellToLatLon(h3Cell string) (float64, float64) {
	cell := h3.CellFromString(h3Cell)
	latLng, err := h3.CellToLatLng(cell)
	if err != nil {
		return 0, 0
	}
	return latLng.Lat, latLng.Lng
}

