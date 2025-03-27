package app

import (
	"encoding/json"
	"log"

	"github.com/vegris/alas-go/shared/application"
	"github.com/vegris/alas-go/shared/token"
)

type config struct {
	HTTPPort       string
	RedisHost      string
	KafkaHost      string
	TokenSecret    []byte
	AllowedSources []string
}

var Config *config

func InitializeConfig() {
	Config = &config{
		HTTPPort:       application.ReadEnvWithFallback("HTTP_PORT", "8080"),
		RedisHost:      application.ReadEnv("REDIS_HOST"),
		KafkaHost:      application.ReadEnv("KAFKA_HOST"),
		TokenSecret:    parseTokenSecret(application.ReadEnv("TOKEN_SECRET")),
		AllowedSources: parseAllowedSources(application.ReadEnv("ALLOWED_SOURCES")),
	}
}

func parseTokenSecret(value string) []byte {
	secret, err := token.DecodeSecret(value)
	if err != nil {
		log.Fatalf("Failed to parse token secret: %v", err)
	}
	return secret
}

func parseAllowedSources(value string) []string {
	var allowedSources []string
	if err := json.Unmarshal([]byte(value), &allowedSources); err != nil {
		log.Fatalf("Failed to parse ALLOWED_SOURCES: %v", err)
	}
	return allowedSources
}
