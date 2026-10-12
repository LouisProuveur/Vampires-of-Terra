package main

import (
	"Vampires-of-Terra/controllers"
	"Vampires-of-Terra/routes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Create context to stop goroutines when stopping the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := controllers.NewHub(ctx)
	router := routes.SetupRouter(hub)

	backendPort := os.Getenv("BACKEND_PORT")
	if backendPort == "" {
		backendPort = "8080" // Default port if not set in .env
	}

	backendAddr := ":" + backendPort
	// Create and start the server with graceful shutdown
	srv := &http.Server{Addr: backendAddr, Handler: router}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
