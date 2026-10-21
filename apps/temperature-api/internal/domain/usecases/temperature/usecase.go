package temperature

import (
	"context"

	"temeperature-service/internal/domain/entities"
)

type UseCase struct {
	temperatureReader Reader
}

func NewUsecase(temperatureReader Reader) UseCase {
	return UseCase{
		temperatureReader: temperatureReader,
	}
}

func (uc *UseCase) GetTemperatureFromLocation(ctx context.Context, location string) (*entities.Temperature, error) {
	return uc.temperatureReader.GetTemperatureByLocation(ctx, location)
}

func (uc *UseCase) GetTemperatureBySensorID(ctx context.Context, id int) (*entities.Temperature, error) {
	return uc.temperatureReader.GetTemperatureByID(ctx, id)
}
