package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
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

type envelope map[string]any

// helper to write json to response
// returns json response body that has been written and the error if there is any
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// use MarshalIndent to add whitespace to encoded json
	// no line prefix and tab indents for each element
	json, err := json.MarshalIndent(data, "", "\t")
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// limit size of request body to 1MB to mitigate DOS on API and prevent larger payloads
	maxBytes := 1_048_576 // 1MB = 1024KB = 1024 * 1024 bytes
	http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// initialise decoder and configure to disallow
	// fields that shouldn't be there
	// decoder will return error if there's an unknown field
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// decode request body into destination (dst any)
	err := dec.Decode(dst)
	if err != nil {
		// error during decoding
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// if syntax error in json body
		// return plain english error message with location of issue
		// example:
		// {"name": "John, "age": 30} -> missing quote
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// Decode() may return io.ErrUnexpectedEOF error
		// for syntax errors in JSON, so check and return readable error
		// example:
		// {"name": "John -> incomplete json
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// if error relates to specific field then
		// include in error message
		// example:
		// {"name": "John", "age": "thirty"} -> age type mismatch
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			// this error is for when json field doesn't exist in struct due to DisallowUnknownFields()
			// example:
			// {"name": "John", "nonexistent_field": "helloworld"} -> nonexistent_field doesn't exist for struct
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// check for when json body is empty and return
		// plain english error message
		// example:
		// "" -> empty body (no JSON)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// panic for invalid unmarshal error
		// decode won't work if destination (dst) is not a pointer
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

			// error for when data in request body is greater than 1MB
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// for anything else, just return error message as is
		default:
			return err
		}
	}

	// call Decode() again using pointer to empty anonymous struct
	// as destination. This is to confirm that request body only confirms
	// a single json value and no additional data
	// this attempts to read more json after first object has been read i.e. {obj1}{obj2}
	// if obj2 exists, its an error, because there should be only 1 obj in reqbody
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	// return no error
	return nil
}

func (app *application) readString(qs url.Values, key, defaultValue string) string {
	// get query param value
	s := qs.Get(key)

	// if no query param value, return default value
	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// get query param value
	csv := qs.Get(key)

	// if no value, return default
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	// try to convert s to int before returning
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be a number")
		return defaultValue
	}

	// return the integer if nothing goes wrong
	return i
}

// helper to run background goroutines
// this helper will manage app waitgroup
func (app *application) background(fn func()) {
	// add to app waitgroup
	app.wg.Add(1)

	// launch background goroutine
	go func() {
		// decrement waitgroup before goroutine returns
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
