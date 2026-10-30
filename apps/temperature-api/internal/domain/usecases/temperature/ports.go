package temperature

import (
	"context"

	"temeperature-service/internal/domain/entities"
)

type Reader interface {
	GetTemperatureByLocation(context.Context, string) (*entities.Temperature, error)
	GetTemperatureByID(context.Context, int) (*entities.Temperature, error)
}
