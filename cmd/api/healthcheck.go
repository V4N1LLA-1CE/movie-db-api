package main

import (
	"net/http"
)

// GET /v1/healthcheck
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// map of response
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process the request", http.StatusInternalServerError)
	}
}
