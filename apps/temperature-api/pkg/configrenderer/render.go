package configrenderer

import (
	"github.com/caarlos0/env/v10"
)

func Render(config any) error {
	return env.Parse(config)
}
