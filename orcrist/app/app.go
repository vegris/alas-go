package app

import (
	"context"
	"embed"
	"net/http"
	"sync"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vegris/alas-go/shared/application"
)

type App struct {
	HTTPRoutes    map[string]http.HandlerFunc
	KafkaHandlers map[string]application.KafkaConsumerHandler
	Jobs          []Job
}

type Job func(context.Context)

type jobContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
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
	Redis = application.StartRedis(Config.RedisHost)
	defer application.ShutdownRedis(Redis)

	DB = application.StartPostgres(Config.PostgresHost, appName, migrationsFS)
	defer application.ShutdownPostgres(DB)

	j := setupJobs(app.Jobs)
	defer cancelJobs(j)

	topicsToCreate := [...]string{KeepAliveTopic, OrcTokensTopic}
	k := application.StartKafka(Config.KafkaHost, topicsToCreate[:], appName, app.KafkaHandlers, Config.KafkaSync)
	Kafka = k.Writer
	defer application.ShutdownKafka(k)

	httpServer := application.StartHTTPServer(Config.HTTPPort, app.HTTPRoutes)
	defer application.ShutdownHTTPServer(httpServer)

	application.BlockUntilInterrupt()
}

func setupJobs(jobs []Job) *jobContext {
	ctx, cancel := context.WithCancel(context.Background())
	c := &jobContext{ctx: ctx, cancel: cancel}

	for _, job := range jobs {
		c.wg.Add(1)
		go startJob(c, job)
	}

	return c
}

func startJob(c *jobContext, j Job) {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			j(c.ctx)
		}
	}
}

func cancelJobs(c *jobContext) {
	c.cancel()
	c.wg.Wait()
}
