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
	Users  UserModel
	// TODO: Add more models here when needed
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
