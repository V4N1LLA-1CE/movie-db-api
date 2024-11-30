package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrUpdateConflict = errors.New("there has been an update conflict")
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
