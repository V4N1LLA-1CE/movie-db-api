package main

import (
	"fmt"
	"net/http"
)

// GET /v1/healthcheck
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: alive")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}
