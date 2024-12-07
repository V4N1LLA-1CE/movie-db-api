package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// declare server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownErr := make(chan error)

	// start background task
	go func() {
		// quit channel to indicate quit
		quit := make(chan os.Signal, 1)

		// listen for incoming SIGINT and SIGTERM and relay to quit channel
		// interrupt: Ctrl+C
		// terminate: kill process i.e. `kill $(lsof -t -i:4000)`
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// wait for signal from quit channel
		// block code execution until signal received
		s := <-quit
		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownErr <- srv.Shutdown(ctx)
	}()

	// start http server
	app.logger.Info("starting server...", "addr", srv.Addr, "env", app.config.env)

	// Shutdown() should immediately cause ListenAndServe() to return
	// http.ErrServerClosed, which indicates graceful shutdown
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// wait to get return from Shutdown()
	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)
	return nil
}
