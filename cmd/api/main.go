package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config config
	logger *slog.Logger
}

func init() {
	// load .env file, otherwise exit
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env filel: %v\n", err)
	}

	err, ok := envExists([]string{
		"POSTGRES_CONTAINER_NAME",
		"POSTGRES_SUPERUSER",
		"POSTGRES_SUPERUSER_PASSWORD",
		"DB_NAME",
		"DB_USER",
		"DB_PASS",
		"PG_DSN",
		"PORT",
		"ENVIRONMENT",
	}...)

	if !ok {
		log.Fatalf(err)
	}
}

func main() {
	// declare config
	var cfg config

	// get env variables into cfg
	cfg.db.dsn = os.Getenv("PG_DSN")

	p, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("failed to parse PORT env, make sure it is a number")
	}
	cfg.port = p

	cfg.env = os.Getenv("ENVIRONMENT")

	// initialise structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// connection pool for db
	conn, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	logger.Info("database connection pool established")

	// declare app
	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start http server
	logger.Info("starting server...", "addr", srv.Addr, "env", cfg.env)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(cfg config) (*sql.DB, error) {
	// create empty connection pool
	conn, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// context with 5-second timeout deadline
	// if max open connections is reached at a time this will make
	// it so User's request will timeout instead of hang indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// establish connection to db
	// if connection cannot be established successfully in 5 seconds
	// there will be an error -> close connection
	err = conn.PingContext(ctx)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func envExists(keys ...string) (string, bool) {
	missing := []string{}

	// loop through all envs and check missing
	for _, key := range keys {
		if e := os.Getenv(key); e == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Sprintln("missing env variables:", strings.Join(missing, ", ")), false
	}

	return "", true
}
