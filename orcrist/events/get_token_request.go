package events

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/vegris/alas-go/shared/schemas"
)

type GetTokenRequest struct {
	EventSource    string      `json:"event_source"`
	EventTimestamp string      `json:"event_timestamp"`
	SessionID      pgtype.UUID `json:"session_id"`
	DeviceInfo     deviceInfo  `json:"device_info"`
}

// deviceInfo represents device-related information
type deviceInfo struct {
	DeviceID           pgtype.UUID `json:"device_id"`
	OS                 string      `json:"os"`
	OSVersion          string      `json:"os_version"`
	DeviceModel        string      `json:"device_model"`
	DeviceManufacturer string      `json:"device_manufacturer"`
}

//go:embed get_token_request.json
var schemaFS embed.FS
var getTokenRequestSchema *jsonschema.Schema

func Initialize() {
	schemaFile, err := schemaFS.Open("get_token_request.json")
	if err != nil {
		log.Fatalf("Failed to open schema file: %v", err)
	}

	schema, err := schemas.CompileSchema(schemaFile)
	if err != nil {
		log.Fatalf("Failed to compile token schema: %v", err)
	}

	getTokenRequestSchema = schema
}

func ParseGetTokenRequest(body []byte) (*GetTokenRequest, error) {
	// Parse into generic interface{} for schema validation
	var requestInstance any
	if err := json.Unmarshal(body, &requestInstance); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate JSON against the schema
	if err := getTokenRequestSchema.Validate(requestInstance); err != nil {
		return nil, fmt.Errorf("invalid GetTokenRequest JSON: %w", err)
	}

	// Parse JSON into a GetTokenRequest struct
	var event GetTokenRequest
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetTokenRequest: %w", err)
	}

	return &event, nil
}
