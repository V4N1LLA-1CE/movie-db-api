package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
	"github.com/google/uuid"
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
	Version   uuid.UUID `json:"version"`           // needed for locking to prevent race conditions
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
	stmt := `INSERT INTO movies (title, year, runtime, genres, version)
  VALUES ($1, $2, $3, $4, uuid_generate_v4())
  RETURNING id, created_at, version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.Genres,
	}

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, created_at, title, year, runtime, genres, version
  FROM movies
  WHERE id = $1`

	var movie Movie

	// create context with timeout
	// 3 seconds after context.WithTimeout() is called, cancel will be called
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// defer cancel() after function to clean up resources (prevents mem leaks)
	// NOTE: - defer will always run at the end of func, regardless of how function exits i.e. early returns etc
	// it should be used for resource cleanup
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(
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
	// use optimistic locking for updating to prevent race conditions
	// https://stackoverflow.com/questions/129329/optimistic-vs-pessimistic-locking/129397#129397
	stmt := `UPDATE movies
  SET title = $1, year = $2, runtime = $3, genres = $4, version = uuid_generate_v4()
  WHERE id = $5 AND version = $6
  RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// execute query with timeout
	// if no matching rows (sql.ErrNoRows), that means the movie version has changed or record has been deleted
	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrUpdateConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	stmt := `DELETE FROM movies
  WHERE id = $1`

	// context with 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// if no rows affected, that means record didn't exist
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// true if title = arg1 or title = ""
	// and genres contains arg2 or genres is empty (which means all)
	stmt := `SELECT id, title, year, runtime, genres, created_at, version
  FROM movies
  WHERE (LOWER(title) = LOWER($1) OR $1 = '')
  AND (genres @> $2 OR $2 = '{}')
  ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, title, genres)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// init slice to hold data
	// use ptrs to pass around movie struct for mem efficiency
	movies := []*Movie{}

	for rows.Next() {
		var movie Movie

		// scan values from row into movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.CreatedAt,
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		// append to movie slice
		movies = append(movies, &movie)
	}

	// when rows.Next() is done, get any errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// if nothing goes wrong, return movie slice
	return movies, nil
}
