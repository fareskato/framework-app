package kabarda

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

////////////////////////////////
// all Kabarda framework routes
////////////////////////////////

// initRoutes create and return chi mux
func (k *Kabarda) initRoutes() http.Handler {
	mux := chi.NewRouter()
	// middlewares
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	// Kabarda custom middlewares
	mux.Use(k.SessionLoad)
	mux.Use(k.NoSurf)
	// only in debug mode
	if k.Debug {
		mux.Use(middleware.Logger)
	}
	// if there is any panic => we recover
	mux.Use(middleware.Recoverer)

	// return the mux(of type http.Handler)
	return mux
}
