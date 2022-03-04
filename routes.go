package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// routes contains all application routes and uses Kabarda Routes of type *chi.Mux
func (a *application) routes() *chi.Mux {
	// middlewares go first

	// routes go here
	a.get("/", a.Handlers.Home)

	// User auth routes
	a.App.Routes.Route("/users", func(r chi.Router) {
		r.Get("/register", a.Handlers.UserRegister)
		r.Post("/register", a.Handlers.UserRegisterPost)
		r.Get("/login", a.Handlers.UserLogin)
		r.Post("/login", a.Handlers.UserLoginPost)
		r.Get("/logout", a.Handlers.UserLogout)
		r.Get("/forgot-password", a.Handlers.ForgotPassword)
		r.Post("/forgot-password", a.Handlers.ForgotPasswordPost)
		r.Get("/reset-password", a.Handlers.ResetPasswordForm)
		r.Post("/reset-password", a.Handlers.ResetPasswordFormPost)
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.handle("/public/*", http.StripPrefix("/public", fileServer))

	/////////////
	// API ROUTES
	/////////////
	a.App.Routes.Mount("/api", a.ApiRoutes())

	// return the mux
	return a.App.Routes
}

///////////////////////////////////////////////
// Helpers functions to make the routing easier
///////////////////////////////////////////////

// get helper
func (a *application) get(path string, h http.HandlerFunc) {
	a.App.Routes.Get(path, h)
}

// post helper
func (a *application) post(path string, h http.HandlerFunc) {
	a.App.Routes.Post(path, h)
}

// patch helper
func (a *application) patch(path string, h http.HandlerFunc) {
	a.App.Routes.Patch(path, h)
}

// put helper
func (a *application) put(path string, h http.HandlerFunc) {
	a.App.Routes.Put(path, h)
}

// delete helper
func (a *application) delete(path string, h http.HandlerFunc) {
	a.App.Routes.Delete(path, h)
}

// handle helper
func (a *application) handle(path string, h http.Handler) {
	a.App.Routes.Handle(path, h)
}

// use helper: for middlewares
func (a *application) use(m ...func(handler http.Handler) http.Handler) {
	a.App.Routes.Use(m...)
}
