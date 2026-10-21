package adapters

import (
	"time"

	"temeperature-service/internal/domain/entities"
)

type LocationRequest struct {
	Location string `form:"location" binding:"required,min=1,max=255"`
}

type SensorIDRequest struct {
	SensorID int `uri:"id" binding:"required,min=1"`
}

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location,omitempty"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id,omitempty"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func ToTemperatureResponse(t *entities.Temperature) TemperatureResponse {
	return TemperatureResponse{
		Value: t.GetTemperature(), Unit: t.GetUnit(), Timestamp: t.GetTimestamp(),
		Location: t.GetLocation(), Status: t.GetStatus(), SensorID: t.GetSensorID(),
		SensorType: t.GetSensorType(), Description: t.GetDescription(),
	}
}
