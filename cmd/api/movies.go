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

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// initialise validator
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// if validation passes, insert into movie db
	err = app.model.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	// add location header so user knows where to find the movie created at /v1/movies/:movieid
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write json response with 201 Created status code along with movie data and headers
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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
