package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vegris/alas-go/kiwi/app"
)

type keepAliveTokensPack struct {
	SessionID uuid.UUID        `json:"session_id"`
	DeviceID  uuid.UUID        `json:"device_id"`
	Tokens    []keepAliveToken `json:"tokens"`
}

type keepAliveToken struct {
	EncodedToken string `json:"encoded"`
	ExpireAt     int64  `json:"expire_at"`
}

var (
	errParseError   = errors.New("Event is malformed")
	errStorageError = errors.New("Storage error")
)

func HandleOrcTokens(message []byte) {
	pack, err := parseKeepAliveTokensPack(message)
	if err != nil {
		return
	}

	if err := storePack(pack); err != nil {
		return
	}

	log.Printf("Sucessfully stored a pack of keep alive tokens!")
}

func parseKeepAliveTokensPack(message []byte) (*keepAliveTokensPack, error) {
	var pack keepAliveTokensPack

	if err := json.Unmarshal(message, &pack); err != nil {
		log.Printf("Failed to unmarshal token pack: %v", err)
		return nil, errParseError
	}

	if len(pack.Tokens) == 0 {
		log.Printf("Received empty keep alive tokens pack")
		return nil, errParseError
	}

	return &pack, nil
}

func storePack(pack *keepAliveTokensPack) error {
	tx := app.Redis.TxPipeline()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	sessionID := pack.SessionID.String()

	members := make([]redis.Z, 0, len(pack.Tokens))
	for _, t := range pack.Tokens {
		member := redis.Z{Score: float64(t.ExpireAt), Member: t.EncodedToken}
		members = append(members, member)
	}

	tx.ZAdd(ctx, sessionID, members...)

	packTTL := calculatePackTTL(pack.Tokens)
	tx.ExpireGT(ctx, sessionID, packTTL)

	if _, err := tx.Exec(ctx); err != nil {
		log.Printf("Failed to store keep alive tokens pack: %v", err)
		return errStorageError
	}

	return nil
}

func calculatePackTTL(tokens []keepAliveToken) time.Duration {
	// Find longest living token
	max := tokens[0].ExpireAt
	for _, value := range tokens {
		if value.ExpireAt > max {
			max = value.ExpireAt
		}
	}

	return time.Unix(max, 0).Sub(time.Now())
}
