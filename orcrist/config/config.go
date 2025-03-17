package config

import (
	"encoding/base64"
	"log"
	"os"
)

type config struct {
	TokenSecret    []byte
	AllowedSources []string
}

var Config *config

func Initialize() {
	Config = &config{
		TokenSecret:    parseTokenSecret(readEnv("TOKEN_SECRET")),
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
	secret, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		log.Fatalf("Failed to decode token secret: %v", err)
	}
	return secret
}

