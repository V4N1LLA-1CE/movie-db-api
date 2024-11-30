package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
	"github.com/lib/pq"
)

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

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// movie model wraps sql.db connection pool
type MovieModel struct {
	DB *sql.DB
}

// CRUD Methods below for movies
func (m MovieModel) Insert(movie *Movie) error {
	stmt := `INSERT INTO movies (title, year, runtime, genres)
  VALUES ($1, $2, $3, $4)
  RETURNING id, created_at, version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.Genres,
	}

	return m.DB.QueryRow(stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, created_at, title, year, runtime, genres, version
  FROM movies
  WHERE id = $1`

	var movie Movie

	err := m.DB.QueryRow(stmt, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres), // can't find pgx equivalent, so now i'm using both pq and pgx, f*ck
		&movie.Version,
	)

	// handle error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	stmt := `UPDATE movies
  SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
  WHERE id = $5
  RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	return m.DB.QueryRow(stmt, args...).Scan(&movie.Version)
}

func (m MovieModel) Delete(id int64) error {
	return nil
}
