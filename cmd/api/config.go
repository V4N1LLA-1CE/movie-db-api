package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

func newConfig() config {
	// declare config
	var cfg config

	// get env variables into cfg
	cfg.db.dsn = os.Getenv("PG_DSN")

	p, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("failed to parse PORT env, is this int type?")
	}

	cfg.port = p
	cfg.env = os.Getenv("ENVIRONMENT")

	moc, err := strconv.Atoi(os.Getenv("POSTGRES_MAXOPENCONNS"))
	if err != nil {
		log.Fatal("failed to parse POSTGRES_MAXOPENCONNS, is this int type?")
	}

	mic, err := strconv.Atoi(os.Getenv("POSTGRES_MAXIDLECONNS"))
	if err != nil {
		log.Fatal("failed to parse POSTGRES_MAXIDLECONNS, is this int type?")
	}

	mit, err := strconv.Atoi(os.Getenv("POSTGRES_MAXIDLETIME"))
	if err != nil {
		log.Fatal("failed to parse POSTGRES_MAXIDLETIME, is this int type?")
	}

	rps, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT"), 64)
	if err != nil {
		log.Fatal("failed to parse RATE_LIMIT, is this float64?")
	}

	burst, err := strconv.Atoi(os.Getenv("RATE_LIMIT_BURST_SIZE"))
	if err != nil {
		log.Fatal("failed to parse RATE_LIMIT_BURST_SIZE, is this int type?")
	}

	rpsEnabled, err := strconv.ParseBool(os.Getenv("RATE_LIMITER_ENABLED"))
	if err != nil {
		log.Fatal("failed to parse RATE_LIMIT_ENABLED, is this bool type?")
	}

	cfg.limiter.rps = rps
	cfg.limiter.burst = burst
	cfg.limiter.enabled = rpsEnabled

	cfg.db.maxOpenConns = moc
	cfg.db.maxIdleConns = mic
	cfg.db.maxIdleTime = time.Duration(mit) * time.Minute

	return cfg
}
