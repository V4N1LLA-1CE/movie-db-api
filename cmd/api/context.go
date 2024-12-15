package main

import (
	"context"
	"net/http"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
)

// custom contextKey type
type contextKey string

const userContextKey = contextKey("user")

// takes a request and user, returns copy of request with the context embedded
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	// create the context with user
	ctx := context.WithValue(r.Context(), userContextKey, user)

	// embed the context into the new copy of request
	return r.WithContext(ctx)
}

// returns the user from the request's context
func (app *application) contextGetUser(r *http.Request) *data.User {
	// get user from contextKey and assert to data.User type
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
