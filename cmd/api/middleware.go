package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
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

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add "Vary: Authorization" header to response
		w.Header().Add("Vary", "Authorization")

		// retrieve value of authorization header from request
		// return "" if no header found
		authorizationHeader := r.Header.Get("Authorization")

		// if no token found in header, set the request context to anonymous user
		// and serve next request
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonUser)
			next.ServeHTTP(w, r)
			return
		}

		// if there is token in auth header,
		// split "Bearer <token>"
		// if wrong format, send 401 unauthorized
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		// validate token
		v := validator.New()

		data.ValidateTokenPlaintext(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// get user associated with token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidCredentialsResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// set user for request context
		r = app.contextSetUser(r, user)

		// go to next request
		next.ServeHTTP(w, r)
	})
}

// middleware for checking if account is authenticated and activated
// use middleware on handlerfunc, not router
func (app *application) requiredActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// check if account is activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// serve following handler
		next.ServeHTTP(w, r)
	})

	// wrap above middleware func with requireAuthenticatedUser() middleware below
	return app.requireAuthenticatedUser(fn)
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnon() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
