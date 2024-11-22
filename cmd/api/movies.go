package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
)

// POST /v1/movies
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// create struct to hold data from post
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// decode request body as json and into input struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// perform validation of data received
	v := validator.New()
	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// dump into http response
	fmt.Fprintf(w, "%+v\n", input)
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
