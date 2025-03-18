package events

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/vegris/alas-go/shared/schemas"
)

// MobileEvent represents the structure of the mobile event
type MobileEvent struct {
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
}

// deviceInfo represents device-related information
type deviceInfo struct {
	DeviceID           uuid.UUID `json:"device_id"`
	OS                 string    `json:"os"`
	OSVersion          string    `json:"os_version"`
	DeviceModel        string    `json:"device_model"`
	DeviceManufacturer string    `json:"device_manufacturer"`
}

// appInfo represents app-related information
type appInfo struct {
	AppVersion     string `json:"app_version"`
	AppBuildNumber string `json:"app_build_number"`
	AppID          string `json:"app_id"`
}

// location represents geolocation data
type location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// networkInfo represents network-related information
type networkInfo struct {
	ConnectionType string `json:"connection_type"`
	Carrier        string `json:"carrier"`
}

//go:embed mobile_event.json
var schemaFS embed.FS
var mobileEventSchema *jsonschema.Schema

func Initialize() {
	schemaFile, err := schemaFS.Open("mobile_event.json")
	if err != nil {
        log.Fatalf("Failed to open schema file: %v", err)
	}

    schema, err := schemas.CompileSchema(schemaFile)
    if err != nil {
        log.Fatalf("Failed to compile token schema: %v", err)
    }

    mobileEventSchema = schema
}

func ParseMobileEvent(body []byte) (*MobileEvent, error) {
	// Parse into generic interface{} for schema validation
	var eventInstance any
	if err := json.Unmarshal(body, &eventInstance); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate JSON against the schema
	if err := mobileEventSchema.Validate(eventInstance); err != nil {
		return nil, fmt.Errorf("invalid mobile event JSON: %w", err)
	}

	// Parse JSON into a MobileEvent struct
	var event MobileEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mobile event: %w", err)
	}

	return &event, nil
}
