package app

import (
	"log"
	"os"

	"github.com/vegris/alas-go/shared/application"
	"github.com/vegris/alas-go/shared/token"
)

type config struct {
	HTTPPort     string
	RedisHost    string
	PostgresHost string
	KafkaHost    string
	KafkaSync    bool
	TokenSecret  []byte
}

var Config *config

func InitializeConfig() {
	Config = &config{
		HTTPPort:     application.ReadEnvWithFallback("HTTP_PORT", "8080"),
		RedisHost:    application.ReadEnv("REDIS_HOST"),
		PostgresHost: application.ReadEnv("POSTGRES_HOST"),
		KafkaHost:    application.ReadEnv("KAFKA_HOST"),
		KafkaSync:    os.Getenv("KAFKA_SYNC") == "1",
		TokenSecret:  parseTokenSecret(application.ReadEnv("TOKEN_SECRET")),
	}
}

func parseTokenSecret(value string) []byte {
	secret, err := token.DecodeSecret(value)
	if err != nil {
		log.Fatalf("Failed to decode token secret: %v", err)
	}
	return secret
}
