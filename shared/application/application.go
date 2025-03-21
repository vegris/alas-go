package application

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type HTTPHandlers = map[string]http.HandlerFunc

func BlockUntilInterrupt() {
	// Listen for SIGTERM to start shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Received stop signal, initiating shutdown")
}

func StartHTTPServer(handlers HTTPHandlers) *http.Server {
	server := &http.Server{Addr: ":8080"}

	for route, handler := range handlers {
		http.HandleFunc(route, handler)
	}

	go func() {
		log.Println("Starting HTTP server on http://localhost:8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	return server
}

func ShutdownHTTPServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err == nil {
		log.Println("HTTP server shutdown successful!")
	} else {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func StartRedis() *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := r.Ping(context.Background()).Err(); err == nil {
		log.Println("Redis initialized")
	} else {
		log.Fatalf("Redis initialization failed: %v", err)
	}

	return r
}

func ShutdownRedis(r *redis.Client) {
	if err := r.Close(); err == nil {
		log.Println("Redis client successfully closed!")
	} else {
		log.Printf("Failed to close Redis client: %v", err)
	}
}
