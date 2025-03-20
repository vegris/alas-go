package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/config"
	"github.com/vegris/alas-go/orcrist/events"
	"github.com/vegris/alas-go/orcrist/handlers"
)

// TODO: extract all app code into shared lib
func main() {
	config.Initialize()
	events.Initialize()

	app.Start(&app.App{
		HTTPRoutes: map[string]http.HandlerFunc{
			"/api/v1/getToken": handlers.HandleGetToken,
		},
		KafkaHandlers: map[string]func([]byte){
			app.KeepAliveTopic: handlers.HandleKeepAlive,
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
