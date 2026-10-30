package adapters

import (
	"math/rand/v2"
	"strconv"
	"time"

	"temeperature-service/internal/domain/entities"
)

func newRandomTemperature(location, sensorID string) *entities.Temperature {
	reading := entities.NewTemperature(
		rand.Float64()*maxTempDeg,
		time.Now().UTC().Add(-time.Duration(rand.Int64N(int64(time.Hour)))),
		location,
		sensorID,
		TemperatureSensorType,
		"Temperature reading",
	)
	return &reading
}

func GetRandomTemperatureByLocation(location string) *entities.Temperature {
	sensorID := "0"
	switch location {
	case "Living Room":
		sensorID = "1"
	case "Bedroom":
		sensorID = "2"
	case "Kitchen":
		sensorID = "3"
	}
	return newRandomTemperature(location, sensorID)
}

func GetRandomTemperatureByID(id int) *entities.Temperature {
	location := "Unknown"
	switch id {
	case 1:
		location = "Living Room"
	case 2:
		location = "Bedroom"
	case 3:
		location = "Kitchen"
	}
	return newRandomTemperature(location, strconv.Itoa(id))
}
