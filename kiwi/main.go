package main

import (
	"net/http"

	"github.com/vegris/alas-go/kiwi/app"
	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/events"
	"github.com/vegris/alas-go/kiwi/handlers"
	"github.com/vegris/alas-go/shared/application"
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
		KafkaHandlers: map[string]application.KafkaConsumerHandler{
			app.OrcTokensTopic: handlers.HandleOrcTokens,
		},
	})
}
