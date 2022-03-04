package main

import (
	"myapp/data"
	"myapp/handlers"
	"myapp/middlewares"

	"github.com/fareskato/kabarda"
)

// application type: wraps Kabarda type and all handlers
type application struct {
	App        *kabarda.Kabarda
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middlewares.Middleware
}

func main() {
	// init(bootstrap) the application
	app := initApplication()
	// Run the server
	app.App.StartServer()
}
