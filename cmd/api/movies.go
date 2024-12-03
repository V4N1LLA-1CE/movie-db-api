package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
	"github.com/google/uuid"
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

	movie, err := app.model.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// write struct ot json and send as http response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// read :id from url param
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// get original movie into a struct
	movie, err := app.model.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// check for X-Expected-Version header, check if movie version
	// in database matches current version specified in the X-Expected-Version header in the
	// current request from client
	if r.Header.Get("X-Expected-Version") != "" {
		expectedVersion, err := uuid.Parse(r.Header.Get("X-Expected-Version"))
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid version format"))
			return
		}
		if movie.Version != expectedVersion {
			app.updateConflictResponse(w, r)
			return
		}
	}

	// hold new data from client
	// use pointers since they have non-zero value
	// if theres no corresponding key in JSON, it will be nil
	// slice already has non zero so no need to use ptrs
	var input struct {
		Title   *string  `json:"title"`
		Year    *int32   `json:"year"`
		Runtime *int32   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// read req body and put data into input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// update movie from Get()
	// this new updated movie struct will be used with Update()
	// since it preserves id and created_at at
	// if input.x is provided, use new values, otherwise just keep it the same (preserve previous)
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// validate updated movie record
	// send 422 Unprocessable Entity response
	// if checks fail
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// update movie
	err = app.model.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUpdateConflict):
			app.updateConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the movie id from the url
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// delete movie from db, send 404 to client if there's no matching record
	err = app.model.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// return 200 with success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// input struct to hold values from request query params
	var input struct {
		Title             string
		Genres            []string
		PaginationOptions data.PaginationOptions
	}

	v := validator.New()

	// get all query params
	// qs is a url.Values type which is a map of query key values pair
	qs := r.URL.Query()

	// read and put values into input data object
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// default is 1 page with 20 size
	input.PaginationOptions.Page = app.readInt(qs, "page", 1, v)
	input.PaginationOptions.PageSize = app.readInt(qs, "page_size", 20, v)

	// default is ascending sort on id
	input.PaginationOptions.Sort = app.readString(qs, "sort", "id")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
