package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/vegris/alas-go/kiwi/token"
)

type OutEvent struct {
	EventName       string                 `json:"event_name"`
	EventType       string                 `json:"event_type"`
	EventSource     string                 `json:"event_source"`
	EventTimestamp  string                 `json:"event_timestamp"`
	UserID          uuid.UUID              `json:"user_id"`
	SessionID       uuid.UUID              `json:"session_id"`
	DeviceInfo      deviceInfo             `json:"device_info"`
	AppInfo         appInfo                `json:"app_info"`
	EventProperties map[string]interface{} `json:"event_properties"`
	Location        location               `json:"location"`
	NetworkInfo     networkInfo            `json:"network_info"`
	ProcessedAt     int64                  `json:"processed_at"`
}

func BuildOutEvent(mobileEvent *MobileEvent, token *token.Token) *OutEvent {
	deviceInfo := mobileEvent.DeviceInfo
	deviceInfo.DeviceID = token.DeviceID

	return &OutEvent{
		EventName:      mobileEvent.EventName,
		EventType:      mobileEvent.EventType,
		EventSource:    mobileEvent.EventSource,
		EventTimestamp: mobileEvent.EventTimestamp,
		UserID:         mobileEvent.UserID,
		// take SessionID and DeviceID from token
		SessionID:       token.SessionID,
		DeviceInfo:      deviceInfo,
		AppInfo:         mobileEvent.AppInfo,
		EventProperties: mobileEvent.EventProperties,
		Location:        mobileEvent.Location,
		NetworkInfo:     mobileEvent.NetworkInfo,
		ProcessedAt:     time.Now().Unix(),
	}
}
