package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// initialise new router
	// use httprouter for OPTIONS
	r := httprouter.New()

	// convert notFoundResponse() helper to a http.handler
	// set it as custom error handler for 404 Not Found responses
	r.NotFound = http.HandlerFunc(app.notFoundResponse)

	// convert helper to http.handler
	// set custom error handler for 405 Method Not Allowed responses
	r.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// healthcheck endpoint
	r.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)

	// movie endpoints
	r.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	r.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	r.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	r.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	r.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// user endpoints
	r.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// return router instance and use middlewares
	return app.recoverPanic(app.rateLimit(r))
}
