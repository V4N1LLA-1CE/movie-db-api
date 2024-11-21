package data

import "time"

// don't use int here to have guaranteed size
type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // hide (not needed in json response)
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`    // don't show in response if empty
	Runtime   int32     `json:"runtime,omitempty"` // don't show in response if empty
	Genres    []string  `json:"genres,omitempty"`  // don't show in response if empty
	Version   int32     `json:"version"`
}
