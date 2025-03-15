package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/handlers"
	"github.com/vegris/alas-go/kiwi/schemas"
)

func main() {
	config.Initialize()
	schemas.Initialize()

	server := setupServer()

    // Listen for SIGTERM to start shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Received stop signal, initiating shutdown")

	shutdownServer(server)
}

func setupServer() *http.Server {
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/api/v1/track", handlers.TrackHandler)

	go func() {
		log.Println("Starting HTTP server on http://localhost:8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	return server
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err == nil {
		log.Println("HTTP server shutdown successful!")
	} else {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
