package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/vegris/alas-go/shared/token"
)

type config struct {
	TokenSecret    []byte
	AllowedSources []string
}

var Config *config

func Initialize() {
	Config = &config{
		TokenSecret:    parseTokenSecret(readEnv("TOKEN_SECRET")),
		AllowedSources: parseAllowedSources(readEnv("ALLOWED_SOURCES")),
	}
}

func readEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("%s environment variable is not set", name)
	}
	return value
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
