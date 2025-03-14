package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vegris/alas-go/kiwi/token"
)

func main() {
	// Initialize the token package
	if err := token.Init(); err != nil {
		log.Fatalf("Failed to initialize token package: %v", err)
	}

	// Create a sample token
	sessionID := uuid.New()
	deviceID := uuid.New()
	expireAt := time.Now().Add(24 * time.Hour).Unix()

	t := token.Token{
		SessionID: sessionID,
		DeviceID:  deviceID,
		ExpireAt:  expireAt,
	}

	// Encode the token
	encoded, err := t.Encode()
	if err != nil {
		log.Fatalf("Failed to encode token: %v", err)
	}
	fmt.Println("Encoded token:", encoded)

	// Decode the token
	decoded, err := token.Decode(encoded)
	if err != nil {
		log.Fatalf("Failed to decode token: %v", err)
	}
	fmt.Printf("Decoded token: %+v\n", decoded)
}
