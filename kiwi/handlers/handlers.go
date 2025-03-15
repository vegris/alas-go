package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/vegris/alas-go/kiwi/events"
	"github.com/vegris/alas-go/kiwi/token"
)

var allowedSources []string

func Init() error {
    envValue := os.Getenv("ALLOWED_SOURCES")
    if envValue == "" {
        return errors.New("ALLOWED_SOURCES environment variable is not set")
    }

    err := json.Unmarshal([]byte(envValue), &allowedSources)
    if err != nil {
        return fmt.Errorf("Failed to parse ALLOWED_SOURCES: %v", err)
    }
    return nil
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

var (
	ErrNoHash             = errors.New("x-hash header is not set")
	ErrHashError          = errors.New("Computed hash did not match")
	ErrNoToken            = errors.New("x-goblin header is not set")
	ErrBadToken           = errors.New("Orc token is invalid")
	ErrReadError          = errors.New("Failed to read request")
	ErrEventError         = errors.New("Event is malformed")
	ErrSourceIsNotAllowed = errors.New("Source is not allowed")
)

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	// Response will always be JSON
	w.Header().Set("Content-Type", "application/json")

	_, err := readOrcToken(r)
	if err != nil {
		handleError(w, err)
		return
	}

	signature, err := readSignature(r)
	if err != nil {
		handleError(w, err)
		return
	}

	body, err := readBody(r)
	if err != nil {
		handleError(w, err)
		return
	}

	event, err := parseEvent(body)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := checkSignature(signature, body, event); err != nil {
		handleError(w, err)
		return
	}

	if err := checkSource(event); err != nil {
		handleError(w, err)
		return
	}

	json.NewEncoder(w).Encode(response{Status: "OK"})
}

func readOrcToken(r *http.Request) (*token.Token, error) {
	header := r.Header.Get("x-goblin")

	// token.Decode can work with empty strings
	t, err := token.Decode(header)
	if err != nil {
		return nil, ErrBadToken
	}
	return t, nil
}

func readSignature(r *http.Request) (string, error) {
	signature := r.Header.Get("x-hash")
	if signature == "" {
		return "", ErrNoHash
	}
	return signature, nil
}

func readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, ErrReadError
	}
	defer r.Body.Close()
	return body, nil
}

func parseEvent(body []byte) (*events.MobileEvent, error) {
	event, err := events.ParseMobileEvent(body)
	if err != nil {
		return nil, ErrEventError
	}
	return event, nil
}

func checkSignature(signature string, body []byte, event *events.MobileEvent) error {
	salt := []byte(event.SessionID.String())
	hash := sha256.Sum256(append(body, salt...))
	hashHexed := hex.EncodeToString(hash[:])

	if signature != hashHexed {
		return ErrHashError
	}
	return nil
}

func checkSource(event *events.MobileEvent) error {
    if !slices.Contains(allowedSources, event.EventSource) {
        return ErrSourceIsNotAllowed
    }
    return nil
}

func handleError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(response{
		Status:  "ERROR",
		Message: err.Error(),
	})
}
