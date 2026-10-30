package stub

import (
	"context"

	"temeperature-service/internal/domain/entities"
	"temeperature-service/internal/infra/stub/adapters"
)

type TemperatureReaderStub struct {
}

func NewTemperatureReader() *TemperatureReaderStub {
	return new(TemperatureReaderStub)
}

func (*TemperatureReaderStub) GetTemperatureByLocation(_ context.Context, location string) (*entities.Temperature, error) {
	return adapters.GetRandomTemperatureByLocation(location), nil
}

func (*TemperatureReaderStub) GetTemperatureByID(_ context.Context, id int) (*entities.Temperature, error) {
	return adapters.GetRandomTemperatureByID(id), nil
}
