package models

type StormType string

const (
	TORNADO StormType = "tornado"
	HAIL    StormType = "hail"
	WIND    StormType = "wind"
)

type StormReport struct {
	Date     string    `json:"date" bson:"date"`
	Time     int32     `json:"time" bson:"time"`
	Size     float64   `json:"size" bson:"size"`
	F_Scale  string    `json:"fScale" bson:"fScale"`
	Speed    int32     `json:"speed" bson:"speed"`
	Location string    `json:"location" bson:"location"`
	County   string    `json:"county" bson:"county"`
	State    string    `json:"state" bson:"state"`
	Lat      float64   `json:"lat" bson:"lat"`
	Lon      float64   `json:"lon" bson:"lon"`
	Comments string    `json:"comments" bson:"comments"`
	Type     StormType `json:"type" bson:"type"`
}

type StormDAOInterface interface {
	GetStormReports(start string, end string) ([]StormReport, error)
	Disconnect() error
}
