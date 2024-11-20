package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// initialise new router
	// use httprouter for OPTIONS
	r := httprouter.New()
	r.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	r.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	r.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// return router instance
	return r
}
