package main

import (
	"context"
	"log"

	"temeperature-service/internal/controller/httpserver"
	"temeperature-service/internal/controller/httpserver/handlers/temperature"
	temperatureUC "temeperature-service/internal/domain/usecases/temperature"
	"temeperature-service/internal/infra/stub"

	"temeperature-service/pkg/app"
	"temeperature-service/pkg/configrenderer"

	"github.com/gin-gonic/gin"
)

func main() {
	var config Config

	if err := configrenderer.Render(&config); err != nil {
		log.Print("failed to render config")

		return
	}

	tempReader := stub.NewTemperatureReader()

	application := app.New(0)

	defer application.Close()

	if err := application.AddController("http listener", func(ctx context.Context) error {
		router := gin.Default()

		var tempertatureUseCase = temperatureUC.NewUsecase(tempReader)

		temperature.NewHandlerRoute(router, tempertatureUseCase)

		return httpserver.NewHTTPServer(router, config.HTTPServer).ListenAndServe(ctx)
	}); err != nil {
		log.Printf("failed to add controller: %s", err)

		return
	}

	if err := application.Run(); err != nil {
		log.Printf("application runtime error: %v", err)
	}
}
