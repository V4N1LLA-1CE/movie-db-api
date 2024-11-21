package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
)

// POST /v1/movies
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// GET /v1/movies/:id
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// read :id from url param
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// dummy data
	movie := envelope{
		"movie": data.Movie{
			ID:        id,
			CreatedAt: time.Now(),
			Title:     "Kimi No Nawa",
			Runtime:   102,
			Genres:    []string{"romance"},
			Version:   1,
		},
	}

	// write struct ot json and send as http response
	err = app.writeJSON(w, http.StatusCreated, movie, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
