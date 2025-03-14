package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrackHandler(t *testing.T) {
	tests := []struct {
		name           string
		xHashHeader    string
		xGoblinHeader  string
		expectedBody   string
	}{
		{
			name:           "Headers match",
			xHashHeader:    "abc123",
			xGoblinHeader:  "abc123",
			expectedBody:   `{"status":"OK"}`,
		},
		{
			name:           "Headers do not match",
			xHashHeader:    "abc123",
			xGoblinHeader:  "xyz789",
			expectedBody:   `{"status":"ERROR","message":"x-hash and x-goblin headers do not match"}`,
		},
		{
			name:           "Missing x-hash header",
			xHashHeader:    "",
			xGoblinHeader:  "xyz789",
			expectedBody:   `{"status":"ERROR","message":"x-hash and x-goblin headers do not match"}`,
		},
		{
			name:           "Missing x-goblin header",
			xHashHeader:    "abc123",
			xGoblinHeader:  "",
			expectedBody:   `{"status":"ERROR","message":"x-hash and x-goblin headers do not match"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request with the POST method and no body
			req, err := http.NewRequest("POST", "/api/v1/track", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			// Set the headers
			req.Header.Set("x-hash", tt.xHashHeader)
			req.Header.Set("x-goblin", tt.xGoblinHeader)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler directly
			TrackHandler(rr, req)

			// Check the response body
			body := strings.TrimSpace(rr.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}
