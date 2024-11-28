package data

import (
	"database/sql"
	"errors"
)

// custom error for when Get() method looks up a movie that doesn't exist
var (
	ErrRecordNotFound = errors.New("record not found")
)

// models struct wraps all models using a single container
type Models struct {
	Movies MovieModel
	// TODO: Add more models here when needed
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
