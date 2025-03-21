package main

import (
	"net/http"

	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/config"
	"github.com/vegris/alas-go/orcrist/events"
	"github.com/vegris/alas-go/orcrist/handlers"
	"github.com/vegris/alas-go/shared/application"
)

func main() {
	config.Initialize()
	events.Initialize()

	app.Start(&app.App{
		HTTPRoutes: map[string]http.HandlerFunc{
			"/api/v1/getToken": handlers.HandleGetToken,
		},
		KafkaHandlers: map[string]application.KafkaConsumerHandler{
			app.KeepAliveTopic: handlers.HandleKeepAlive,
		},
	})
}
