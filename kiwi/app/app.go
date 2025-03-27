package app

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vegris/alas-go/shared/application"
)

type App struct {
	HTTPRoutes    map[string]http.HandlerFunc
	KafkaHandlers map[string]application.KafkaConsumerHandler
}

var Redis *redis.Client
var Kafka *kafka.Writer

const appName = "kiwi"

const EventsTopic = "kiwi-events"
const KeepAliveTopic = "keep-alive"
const OrcTokensTopic = "orc-tokens"

func Start(app *App) {
	Redis = application.StartRedis(Config.RedisHost)
	defer application.ShutdownRedis(Redis)

	topicsToCreate := [...]string{EventsTopic, KeepAliveTopic, OrcTokensTopic}
	k := application.StartKafka(Config.KafkaHost, topicsToCreate[:], appName, app.KafkaHandlers)
	Kafka = k.Writer
	defer application.ShutdownKafka(k)

	httpServer := application.StartHTTPServer(Config.HTTPPort, app.HTTPRoutes)
	defer application.ShutdownHTTPServer(httpServer)

	application.BlockUntilInterrupt()
}
