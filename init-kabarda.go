package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"myapp/middlewares"
	"os"

	"github.com/fareskato/kabarda"
)

// initApplication bootstraps the application
func initApplication() *application {
	// get working dir
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// init Kabarda
	kbr := &kabarda.Kabarda{}
	// call New on kbr
	err = kbr.New(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	// populate kbr fields
	kbr.AppName = "myapp"

	// middlewares
	appMiddleware := &middlewares.Middleware{
		App: kbr,
	}

	// handlers
	appHandlers := &handlers.Handlers{
		App: kbr,
	}

	// create app
	app := &application{
		App:        kbr,
		Handlers:   appHandlers,
		Middleware: appMiddleware,
	}

	// add all the app routes to Kabarda routes
	app.App.Routes = app.routes()

	// init all models which is Models Type: package data Models type
	app.Models = data.New(app.App.DB.Pool)

	// add models to handler so all handlers can access all models
	appHandlers.Models = app.Models

	// add models to the middleware
	app.Middleware.Models = app.Models

	// return the app
	return app

}
