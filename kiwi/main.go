package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/events"
	"github.com/vegris/alas-go/kiwi/handlers"
	"github.com/vegris/alas-go/kiwi/token"
)

func main() {
    config.Initialize()

	if err := token.Init(); err != nil {
		log.Fatalf("Failed to initialize token package: %v", err)
	}

    checkTokenPackage()

	if err := events.Init(); err != nil {
		log.Fatalf("Failed to initialize events package: %v", err)
	}

	// Register the trackHandler function to handle requests at /api/v1/track
	http.HandleFunc("/api/v1/track", handlers.TrackHandler)

	// Start the HTTP server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func checkTokenPackage() {
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
