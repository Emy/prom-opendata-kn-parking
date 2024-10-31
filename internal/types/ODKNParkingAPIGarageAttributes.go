package types

type ODKNParkingAPIGarageAttributes struct {
	ObjectID int     `json:"OBJECTID"`
	ID       float64 `json:"id"`
	Name     string  `json:"name"`
	MaxCap   float64 `json:"max_cap"`
	Type     string  `json:"type"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	RealFCap *int    `json:"real_fcap"` // Use pointer to handle null values
	RealCapa int     `json:"real_capa"`
}
