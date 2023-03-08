package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"viadro_api/internal/logger"
)

func (app *application) serve(app_port string) error {
	srv := http.Server{
		//add TLS config
		Addr:         fmt.Sprintf(":%s", app_port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	shutdown_signal := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.LogInfo(fmt.Sprintf("shutdown signal: %s", s.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdown_signal <- err
		}

		logger.LogInfo("waiting for background tasks to finish")

		app.wait_group.Wait()
		shutdown_signal <- nil
	}()

	logger.LogInfo("starting server")
	err := srv.ListenAndServe()
	if err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			return nil
		default:
			return err
		}
	}

	err = <-shutdown_signal
	if err != nil {
		return err
	}

	return nil
}
