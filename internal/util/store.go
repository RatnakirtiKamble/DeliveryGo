package util

type StoreLocation struct {
	Lat float64
	Lon float64
}

var StoreLocations = map[string]StoreLocation{
	"store-1": {
		Lat: 19.104626,
		Lon: 73.003936, 
	},
}
