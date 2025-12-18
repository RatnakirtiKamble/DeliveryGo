package domain

type PathTemplate struct {
	ID			string 
	StoreID		string
	Polyline 	[]GeoPoint 
	BaseETA 	int 
}

type GeoPoint struct{
	Lat float64 
	Lon float64
}