package middlewares

import (
	"myapp/data"

	"github.com/fareskato/kabarda"
)

type Middleware struct {
	App    *kabarda.Kabarda
	Models data.Models
}
