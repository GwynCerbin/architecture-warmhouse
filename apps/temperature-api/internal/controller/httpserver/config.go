package httpserver

import "time"

type Config struct {
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"`
	Port            int           `env:"PORT"`
}
