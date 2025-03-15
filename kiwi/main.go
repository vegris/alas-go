package main

import (
	"fmt"
	"net/http"

	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/schemas"
	"github.com/vegris/alas-go/kiwi/handlers"
)

func main() {
    config.Initialize()
    schemas.Initialize()

	// Register the trackHandler function to handle requests at /api/v1/track
	http.HandleFunc("/api/v1/track", handlers.TrackHandler)

	// Start the HTTP server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
