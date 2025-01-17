package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/data"
	"github.com/V4N1LLA-1CE/movie-db-api/internal/mailer"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func init() {
	// load .env file, otherwise exit
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env filel: %v\n", err)
	}

	err, ok := envExists([]string{
		"POSTGRES_CONTAINER_NAME",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB_NAME",
		"PG_DSN",
		"PORT",
		"ENVIRONMENT",
		"POSTGRES_MAXOPENCONNS",
		"POSTGRES_MAXIDLECONNS",
		"POSTGRES_MAXIDLETIME",
		"RATE_LIMITER_ENABLED",
		"RATE_LIMIT",
		"RATE_LIMIT_BURST_SIZE",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_SENDER",
	}...)

	if !ok {
		log.Fatal(err)
	}
}

func main() {
	// get app configuration
	cfg := newConfig()

	// get trusted origins from cli flag
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	flag.Parse()
	if len(cfg.cors.trustedOrigins) == 0 {
		log.Fatal("Need cors-trusted-origins flag value")
	}

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
		models: data.NewModels(conn),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// start server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	// create empty connection pool
	conn, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// set the following configs for db:
	// max number of open (in-use + idle) connections in pool
	// max number of idle connections in pool
	// maximum timeout for idle connections (conns not being used)
	conn.SetMaxOpenConns(cfg.db.maxOpenConns)
	conn.SetMaxIdleConns(cfg.db.maxIdleConns)
	conn.SetConnMaxIdleTime(cfg.db.maxIdleTime)

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
