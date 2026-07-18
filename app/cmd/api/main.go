package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scan/internal/config"
	"scan/internal/handlers"
	"scan/internal/router"
	"scan/internal/storage"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.Open(cfg.BadgerDBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	log.Printf("Starting server on port %s", cfg.ServerPort)
	h := handlers.NewHandlers(cfg, db)
	mux := router.New(h)
	serverAddr := ":" + cfg.ServerPort

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server: listening on %s", serverAddr)
		serverErr <- server.ListenAndServe()
	}()

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("server error: %v", err)
	case sig := <-sigChan:
		log.Printf("server: received signal %v, initiating graceful shutdown", sig)
	}

	shutdownTimeout := 30 * time.Second
	log.Println("server: shutting down...")

	log.Println("server: closing database...")
	if err := db.Close(); err != nil {
		log.Printf("server: database close error: %v", err)
	}

	log.Println("server: shutting down HTTP server...")
	ctx, cancel := context.WithTimeout(shutdownCtx, shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server: HTTP shutdown error: %v", err)
	}

	log.Println("server: shutdown complete")
}
