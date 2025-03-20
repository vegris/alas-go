package config

import (
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
		TokenSecret: parseTokenSecret(readEnv("TOKEN_SECRET")),
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
		log.Fatalf("Failed to decode token secret: %v", err)
	}
	return secret
}
