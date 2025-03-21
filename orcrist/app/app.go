package app

import (
	"embed"
	"net/http"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vegris/alas-go/shared/application"
)

type App struct {
	HTTPRoutes    map[string]http.HandlerFunc
	KafkaHandlers map[string]application.KafkaConsumerHandler
}

var Redis *redis.Client
var DB *pgxpool.Pool
var Kafka *kafka.Writer

const appName = "orcrist"

const KeepAliveTopic = "keep-alive"
const OrcTokensTopic = "orc-tokens"

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Start(app *App) {
	Redis = application.StartRedis()
	defer application.ShutdownRedis(Redis)

	DB = application.StartPostgres(appName, migrationsFS)
	defer application.ShutdownPostgres(DB)

	topicsToCreate := [...]string{KeepAliveTopic, OrcTokensTopic}
	k := application.StartKafka(topicsToCreate[:], appName, app.KafkaHandlers)
	Kafka = k.Writer
	defer application.ShutdownKafka(k)

	httpServer := application.StartHTTPServer(app.HTTPRoutes)
	defer application.ShutdownHTTPServer(httpServer)

	application.BlockUntilInterrupt()
}
