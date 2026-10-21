package entities

import "time"

type Temperature struct {
	// value is measured in degrees Celsius.
	value       float64
	unit        string
	timestamp   time.Time
	location    string
	status      string
	sensorID    string
	sensorType  string
	description string
}

func (t Temperature) GetTemperature() float64 { return t.value }
func (t Temperature) GetUnit() string         { return t.unit }
func (t Temperature) GetTimestamp() time.Time { return t.timestamp }
func (t Temperature) GetLocation() string     { return t.location }
func (t Temperature) GetStatus() string       { return t.status }
func (t Temperature) GetSensorID() string     { return t.sensorID }
func (t Temperature) GetSensorType() string   { return t.sensorType }
func (t Temperature) GetDescription() string  { return t.description }

func NewTemperature(value float64, timestamp time.Time, location, sensorID, sensorType, description string) Temperature {
	return Temperature{
		value:       value,
		unit:        "°C",
		timestamp:   timestamp.UTC(),
		location:    location,
		status:      "active",
		sensorID:    sensorID,
		sensorType:  sensorType,
		description: description,
	}
}
