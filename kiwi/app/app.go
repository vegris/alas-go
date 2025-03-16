package app

import (
	"context"
	"log"
	"net/http"
	"time"
)

type App struct {
	HTTPRoutes map[string]http.HandlerFunc
}

var httpServer *http.Server

func Start(app *App) {
	startHTTPServer(app)
}

func Shutdown() {
    shutdownHTTPServer()
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
