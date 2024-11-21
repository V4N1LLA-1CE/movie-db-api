package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// helper to read ":id" url parameters
// i.e. /v1/movies/:id
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	// parse id into int64
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// helper to write json to response
// returns json response body that has been written and the error if there is any
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// better formatting for terminal output i.e. curl
	json = append(json, '\n')

	// loop through header map and set key and value
	for hkey, hval := range headers {
		w.Header()[hkey] = hval
	}

	// set json header and status code
	// then write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}
