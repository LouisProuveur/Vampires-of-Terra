// src/main.go
package main

import (
	"Vampires-of-Terra/controllers"
	"Vampires-of-Terra/models"
	"Vampires-of-Terra/routes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initLogging() {
	level := slog.LevelInfo

	switch lvlstr := os.Getenv("LOG_LEVEL"); lvlstr {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	default:
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

func main() {

	initLogging()

	// Create context to stop goroutines when stopping the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := models.RegisterValidators(); err != nil {
		slog.Error("validator loading failed", "error", err)
		os.Exit(1)
	}

	hub := controllers.NewHub(ctx, os.Getenv("FRONTEND_URL"))

	router := routes.SetupRouter(hub)

	backendPort := os.Getenv("BACKEND_PORT")
	if backendPort == "" {
		backendPort = "8080" // Default port if not set in env
	}

	backendAddr := ":" + backendPort
	// Create and start the server with graceful shutdown
	srv := &http.Server{
		Addr:              backendAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		MaxHeaderBytes:    1 << 20,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("error when stopping server", "error", err)
		}
		close(done)
		slog.Info("server stopped")
	}()

	slog.Info("server started", "port", backendPort)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("error while starting server", "error", err)
		os.Exit(1)
	}
	<-done
}

// TODO: better validator messages, rather than dump err.Error()
// TODO: use env variables when possible rather const when necessary
// TODO: document the code, and modify existing doc comments to match Go convention (especially exported methods)
// TODO: change filepath comments
