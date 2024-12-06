package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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

// NOTE: This method of rate limiting will only work if API is hosted on a single machine
// If infra is distributed with multiple servers and load balancer, it won't work
func (app *application) rateLimit(next http.Handler) http.Handler {
	// client struct for holding rate limiter and last seen time
	type client struct {
		limiter           *rate.Limiter
		mostRecentRequest time.Time
	}

	// map of clients that get mapped by client IP addresses
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// background goroutine to cleanup clients every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock for cleanup
			mu.Lock()

			// if client hasn't requested for 3 minutes, delete from map
			for ip, client := range clients {
				if time.Since(client.mostRecentRequest) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// unlock mutex
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only carry out check if rate limiting is enabled
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// lock mutex to allow only 1 goroutine to read/write to clients map at a time
			mu.Lock()

			// check if ip exists in map, if not, add a new client and add to map with ip
			if _, exists := clients[ip]; !exists {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			// update client's most recent request
			clients[ip].mostRecentRequest = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// unlock mutex
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}
