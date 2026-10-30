package temperature

import (
	"net/http"

	"temeperature-service/internal/controller/httpserver/handlers/temperature/adapters"
	"temeperature-service/internal/domain/entities"
	"temeperature-service/internal/domain/usecases/temperature"
	"temeperature-service/pkg/ginbinder"

	"github.com/gin-gonic/gin"
)

type handlerSet struct {
	usecase temperature.UseCase
}

func NewHandlerRoute(router *gin.Engine, usecase temperature.UseCase) {
	var (
		hs = handlerSet{usecase: usecase}
		rg = router.Group("temperature")
	)

	rg.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, ginbinder.Response{
			Success: true,
			Message: "ok",
		})
	})

	rg.GET("/:id", hs.GetTemperatureByID)
	rg.GET("", hs.GetTemperatureByLocation)
}

func (hs *handlerSet) GetTemperatureByLocation(c *gin.Context) {
	var request adapters.LocationRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(ginbinder.HandleBinding(err))

		return
	}

	reading, err := hs.usecase.GetTemperatureFromLocation(c.Request.Context(), request.Location)

	respond(c, reading, err)
}

func (hs *handlerSet) GetTemperatureByID(c *gin.Context) {
	var request adapters.SensorIDRequest
	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(ginbinder.HandleBinding(err))

		return
	}

	reading, err := hs.usecase.GetTemperatureBySensorID(c.Request.Context(), request.SensorID)

	respond(c, reading, err)
}

func respond(c *gin.Context, reading *entities.Temperature, err error) {
	if err != nil || reading == nil {
		c.JSON(http.StatusInternalServerError, ginbinder.Response{Message: "Не удалось получить данные датчика"})
		return
	}

	c.JSON(http.StatusOK, adapters.ToTemperatureResponse(reading))
}
