package handlers

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	// Extract headers
	hashHeader := r.Header.Get("x-hash")
	goblinHeader := r.Header.Get("x-goblin")

	var resp response
	// Compare the headers
	if hashHeader == goblinHeader {
		// If headers are equal, return OK response
		resp = response{Status: "OK"}
	} else {
		// If headers are not equal, return ERROR response
		resp = response{
			Status:  "ERROR",
			Message: "x-hash and x-goblin headers do not match",
		}
	}

	// Set the Content-Type to application/json
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(resp)
}
