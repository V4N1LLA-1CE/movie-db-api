package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
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
}

func main() {
	// declare config
	var cfg config

	// get postgres dsn from env
	cfg.db.dsn = os.Getenv("PG_DSN")

	// read port and env command line flags and write into config
	flag.IntVar(&cfg.port, "p", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev|staging|prod)")
	flag.Parse()

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
