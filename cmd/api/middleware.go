package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	// initialise new rate limiter
	// rate limit: 2 - on average allows 2 events/operations per second
	// burst size: 4 - max number of tokens in bucket (4 tokens max so 4 requests in quick succession)
	limiter := rate.NewLimiter(2, 4)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// everytime Allow() is called, 1 token will be consumed from bucket
		// if no tokens, Allow() is false
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
