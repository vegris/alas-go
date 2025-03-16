package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vegris/alas-go/kiwi/app"
	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/events"
	"github.com/vegris/alas-go/kiwi/token"
)

type okResponse struct {
	Status   string `json:"status"`
	Token    string `json:"token"`
	TokenTTL int64  `json:"ttl"`
}

type errResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	errNoHash             = errors.New("x-hash header is not set")
	errHashError          = errors.New("Computed hash did not match")
	errNoToken            = errors.New("x-goblin header is not set")
	errBadToken           = errors.New("Orc token is invalid")
	errReadError          = errors.New("Failed to read request")
	errEventError         = errors.New("Event is malformed")
	errSourceIsNotAllowed = errors.New("Source is not allowed")
	errNoFreshToken       = errors.New("Failed to refresh Orc token")
	errInternalError      = errors.New("Internal server error")
)

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	// Response will always be JSON
	w.Header().Set("Content-Type", "application/json")

	oldToken, err := readOrcToken(r)
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

	freshToken, tokenTTL, err := refreshToken(oldToken)
	if err != nil {
		handleError(w, err)
		return
	}

    if err := produceKeepAliveEvent(event, oldToken); err != nil {
       handleError(w, err) 
       return
    }

	json.NewEncoder(w).Encode(okResponse{Status: "OK", Token: freshToken, TokenTTL: tokenTTL})
}

func readOrcToken(r *http.Request) (*token.Token, error) {
	header := r.Header.Get("x-goblin")

	// token.Decode can work with empty strings
	t, err := token.Decode(header)
	if err != nil {
		return nil, errBadToken
	}
	return t, nil
}

func readSignature(r *http.Request) (string, error) {
	signature := r.Header.Get("x-hash")
	if signature == "" {
		return "", errNoHash
	}
	return signature, nil
}

func readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errReadError
	}
	defer r.Body.Close()
	return body, nil
}

func parseEvent(body []byte) (*events.MobileEvent, error) {
	event, err := events.ParseMobileEvent(body)
	if err != nil {
		return nil, errEventError
	}
	return event, nil
}

func checkSignature(signature string, body []byte, event *events.MobileEvent) error {
	salt := []byte(event.SessionID.String())
	hash := sha256.Sum256(append(body, salt...))
	hashHexed := hex.EncodeToString(hash[:])

	if signature != hashHexed {
		return errHashError
	}
	return nil
}

func checkSource(event *events.MobileEvent) error {
	if !slices.Contains(config.Config.AllowedSources, event.EventSource) {
		return errSourceIsNotAllowed
	}
	return nil
}

func refreshToken(token *token.Token) (string, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	now := time.Now().Unix()
	minTokenTTL := now + int64(time.Minute)
	maxTokenTTL := now + int64(time.Minute)*2

	zrange := &redis.ZRangeBy{Min: strconv.FormatInt(minTokenTTL, 10), Max: strconv.FormatInt(maxTokenTTL, 10), Offset: 0, Count: 1}

	vals, err := app.Redis.ZRangeByScoreWithScores(ctx, token.SessionID.String(), zrange).Result()
	if err != nil {
		log.Printf("Error accessing Redis: %v", err)
		return "", 0, errInternalError
	}
	if len(vals) == 0 {
		return "", 0, errNoFreshToken
	}

	encodedToken := vals[0].Member.(string)
	tokenExpireAt := int64(vals[0].Score)

	ttl := tokenExpireAt - now

	return encodedToken, ttl, nil
}

func produceKeepAliveEvent(mobileEvent *events.MobileEvent, token *token.Token) error {
    outEvent := events.BuildKeepAliveEvent(mobileEvent, token)
    outMessage, err := json.Marshal(outEvent)
    if err != nil {
        // This should never happen
        log.Fatalf("Encoding KeepAlive event to JSON failed: %v", err)
    }

    message := kafka.Message{
        Topic: app.KeepAliveTopic,
        Value: outMessage,
    }

    if err := app.Kafka.WriteMessages(context.Background(), message); err != nil {
		log.Printf("Error producing to Redis: %v", err)
        return errInternalError
    }

    return nil
}

func handleError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(errResponse{
		Status:  "ERROR",
		Message: err.Error(),
	})
}
