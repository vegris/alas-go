package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	HTTPRoutes map[string]http.HandlerFunc
}

var Redis *redis.Client
var httpServer *http.Server

func Start(app *App) {
	startRedis()
	startHTTPServer(app)
}

func Shutdown() {
	shutdownHTTPServer()
	shutdownRedis()
}

func startHTTPServer(app *App) {
	httpServer = &http.Server{Addr: ":8080"}

	for route, handler := range app.HTTPRoutes {
		http.HandleFunc(route, handler)
	}

	go func() {
		log.Println("Starting HTTP server on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func shutdownHTTPServer() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err == nil {
		log.Println("HTTP server shutdown successful!")
	} else {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func startRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := Redis.Ping(context.Background()).Err(); err == nil {
		log.Println("Redis initialized")
	} else {
        log.Fatalf("Redis initialization failed: %v", err)
	}
}

func shutdownRedis() {
    if err := Redis.Close(); err == nil {
        log.Println("Redis client successfully closed!")
    } else {
        log.Printf("Failed to close Redis client: %v", err)
    }
}
