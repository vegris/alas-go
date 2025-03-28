package main

import (
	"time"

	"github.com/google/uuid"
)

type MobileEvent struct {
	EventName       string                 `json:"event_name"`
	EventType       string                 `json:"event_type"`
	EventSource     string                 `json:"event_source"`
	EventTimestamp  string                 `json:"event_timestamp"`
	UserID          uuid.UUID              `json:"user_id"`
	SessionID       uuid.UUID              `json:"session_id"`
	DeviceInfo      DeviceInfo             `json:"device_info"`
	AppInfo         AppInfo                `json:"app_info"`
	EventProperties map[string]interface{} `json:"event_properties"`
	Location        Location               `json:"location"`
	NetworkInfo     NetworkInfo            `json:"network_info"`
}

type GetTokenRequest struct {
	EventSource    string     `json:"event_source"`
	EventTimestamp string     `json:"event_timestamp"`
	SessionID      uuid.UUID  `json:"session_id"`
	DeviceInfo     DeviceInfo `json:"device_info"`
}

type DeviceInfo struct {
	DeviceID           uuid.UUID `json:"device_id"`
	OS                 string    `json:"os"`
	OSVersion          string    `json:"os_version"`
	DeviceModel        string    `json:"device_model"`
	DeviceManufacturer string    `json:"device_manufacturer"`
}

type AppInfo struct {
	AppVersion     string `json:"app_version"`
	AppBuildNumber string `json:"app_build_number"`
	AppID          string `json:"app_id"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type NetworkInfo struct {
	ConnectionType string `json:"connection_type"`
	Carrier        string `json:"carrier"`
}

func mobileEvent() *MobileEvent {
	return &MobileEvent{
		EventName:      "screen_view",
		EventType:      "navigation",
		EventSource:    "mobile_app",
		EventTimestamp: time.Now().UTC().Format(time.RFC3339),
		UserID:         uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		SessionID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		DeviceInfo: DeviceInfo{
			DeviceID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			OS:                 "iOS",
			OSVersion:          "15.4",
			DeviceModel:        "iPhone 13",
			DeviceManufacturer: "Apple",
		},
		AppInfo: AppInfo{
			AppVersion:     "2.3.0",
			AppBuildNumber: "20300",
			AppID:          "com.example.myapp",
		},
		EventProperties: map[string]interface{}{
			"screen_name":             "Home",
			"previous_screen":         "Splash",
			"time_on_previous_screen": 2.5,
		},
		Location: Location{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
		NetworkInfo: NetworkInfo{
			ConnectionType: "wifi",
			Carrier:        "AT&T",
		},
	}
}

func getTokenRequest() *GetTokenRequest {
	return &GetTokenRequest{
		EventSource:    "mobile_app",
		EventTimestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		DeviceInfo: DeviceInfo{
			DeviceID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			OS:                 "iOS",
			OSVersion:          "15.4",
			DeviceModel:        "iPhone 13",
			DeviceManufacturer: "Apple",
		},
	}
}
