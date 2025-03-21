package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/repo"
	"github.com/vegris/alas-go/shared/token"
)

type keepAliveEvent struct {
	EventName      string      `json:"event_name"`
	EventType      string      `json:"event_type"`
	EventSource    string      `json:"event_source"`
	EventTimestamp string      `json:"event_timestamp"`
	UserID         pgtype.UUID `json:"user_id"`
	SessionID      pgtype.UUID `json:"session_id"`
	DeviceInfo     deviceInfo  `json:"device_info"`
	ProcessedAt    int64       `json:"processed_at"`
}

type deviceInfo struct {
	DeviceID           pgtype.UUID `json:"device_id"`
	OS                 string      `json:"os"`
	OSVersion          string      `json:"os_version"`
	DeviceModel        string      `json:"device_model"`
	DeviceManufacturer string      `json:"device_manufacturer"`
}

var errSessionNotFound = errors.New("Session not found")

func HandleKeepAlive(message []byte) {
	var event *keepAliveEvent
	if err := json.Unmarshal(message, event); err != nil {
		log.Printf("Failed to decode keep alive event: %v", err)
		return
	}

	err := prolongSession(event.SessionID, event.ProcessedAt)
	if errors.Is(err, errSessionNotFound) {
		log.Printf("Session not found: %v", event.SessionID)
		return
	}
	if err != nil {
		// Generate future tokens even if prolonging session failed
		log.Printf("Failed to prolong session %v: %v", event.SessionID, err)
	}

	token := token.Token{SessionID: event.SessionID.Bytes, DeviceID: event.DeviceInfo.DeviceID.Bytes, ExpireAt: event.ProcessedAt}

	if err := generateFutureTokens(event.SessionID, &token); err != nil {
		log.Printf("Failed to generate future tokens: %v", err)
		return
	}
}

func prolongSession(sessionID pgtype.UUID, eventTS int64) error {
	q := repo.New(app.DB)
	ctx := context.Background()

	session, err := q.GetAliveSession(ctx, sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errSessionNotFound
	}
	if err != nil {
		return err
	}

	if session.EndsAt.Time.Unix() >= eventTS {
		return nil
	}

	params := repo.RefreshSessionParams{SessionID: sessionID, SessionDuration: eventTS}
	session, err = q.RefreshSession(ctx, params)

	return err
}
