package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *application) ApiRoutes() http.Handler {
	r := chi.NewRouter()

	// use Route to group routes
	// example
	/*
		r.Route("/todos", func(r chi.Router) {
			r.Get("/show", func(w http.ResponseWriter, r *http.Request) {
				var todo struct {
					Name string `json:"name"`
				}
				todo.Name = "do coding"
				a.App.WriteJSON(w, 200, todo)
			})
		})
	*/
	return r
}
