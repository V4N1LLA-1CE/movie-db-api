package main

import (
	"net/http"
)

// GET /v1/healthcheck
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// map of response
	health := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, health, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
