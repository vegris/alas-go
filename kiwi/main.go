package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vegris/alas-go/kiwi/app"
	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/events"
	"github.com/vegris/alas-go/kiwi/handlers"
	"github.com/vegris/alas-go/shared/token"
)

func main() {
	config.Initialize()
	token.Initialize()
	events.Initialize()

	app.Start(&app.App{
		HTTPRoutes: map[string]http.HandlerFunc{
			"/api/v1/track": handlers.TrackHandler,
		},
		KafkaHandlers: map[string]func([]byte){
			app.OrcTokensTopic: handlers.HandleOrcTokens,
		},
	})

	waitStop()

	app.Shutdown()
}

func waitStop() {
	// Listen for SIGTERM to start shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Received stop signal, initiating shutdown")
}
