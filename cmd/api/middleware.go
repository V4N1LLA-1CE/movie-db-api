package main

import (
	"fmt"
	"net/http"
)

// middleware allows app to run even during a panic so server doesn't crash
// respond with 500 Internal Server Error and close connection with headers
// if no errors, proceed to serve the handler
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create deferred function that will always run in the event of a panic
		defer func() {
			// built in recover checks for panic
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
