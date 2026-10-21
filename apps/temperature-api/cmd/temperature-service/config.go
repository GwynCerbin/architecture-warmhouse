package main

import (
	"temeperature-service/internal/controller/httpserver"
)

type Config struct {
	HTTPServer httpserver.Config `envPrefix:"HTTP_SERVER_"`
}
