package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/vegris/alas-go/integration_tests/repo"
	"github.com/vegris/alas-go/shared/application"
	"github.com/vegris/alas-go/shared/token"
)

const postgresURL = "postgres://postgres:postgres@localhost:5432/orcrist"

const kafkaAddr = "localhost:9092"
const consumerGroup = "integration-tests"

const trackEventURL = "http://localhost:4000/api/v1/track"
const getTokenURL = "http://localhost:4001/api/v1/getToken"

var tokenSecret []byte
var allowedSources []string

var Kafka *kafka.Reader
var DB *pgx.Conn

func TestMain(m *testing.M) {
	tokenSecret = parseTokenSecret(application.ReadEnv("TOKEN_SECRET"))
	allowedSources = parseAllowedSources(application.ReadEnv("ALLOWED_SOURCES"))

	token.Initialize()

	ctx := context.Background()

	db, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		log.Fatalf("Unable to connect to Postgres: %v", err)
	}
	defer db.Close(ctx)
	DB = db

	Kafka = kafka.NewReader(kafka.ReaderConfig{Brokers: []string{kafkaAddr}, GroupID: consumerGroup, Topic: "kiwi-events", StartOffset: kafka.LastOffset})

	// For some reason kafka-go does not fully initialize reader on NewReader call
	// resulting in first ReadMessage call failing
	// Force full initialization with CommitMessages call
	if err := Kafka.CommitMessages(context.Background()); err != nil {
		log.Printf("Failed to commit offsets to Kafka: %v", err)
	}
	defer Kafka.Close()

	m.Run()
}

func parseTokenSecret(value string) []byte {
	secret, err := token.DecodeSecret(value)
	if err != nil {
		log.Fatalf("Failed to parse token secret: %v", err)
	}
	return secret
}

func parseAllowedSources(value string) []string {
	var allowedSources []string
	if err := json.Unmarshal([]byte(value), &allowedSources); err != nil {
		log.Fatalf("Failed to parse ALLOWED_SOURCES: %v", err)
	}
	return allowedSources
}

func makeRequest(url string, headers map[string]string, payload []byte) map[string]interface{} {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	must(err)

	request.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	r, err := http.DefaultClient.Do(request)
	must(err)
	defer r.Body.Close()

	bytes, err := io.ReadAll(r.Body)
	must(err)

	var response map[string]interface{}
	must(json.Unmarshal(bytes, &response))

	return response
}

func calculateSignature(sessionID uuid.UUID, body []byte) string {
	salt := []byte(sessionID.String())
	hash := sha256.Sum256(append(body, salt...))
	return hex.EncodeToString(hash[:])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func toUUID(pgUUID pgtype.UUID) uuid.UUID {
	return uuid.UUID(pgUUID.Bytes)
}

func toPgUUID(uid uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: uid, Valid: true}
}

// Kiwi tests

func TestKiwiBadSignature(t *testing.T) {
	tokenRequest, err := json.Marshal(getTokenRequest())
	must(err)

	response := makeRequest(getTokenURL, nil, tokenRequest)
	assert.Equal(t, response["status"], "OK")

	token := response["token"].(string)
	assert.NotEmpty(t, token)

	event := mobileEvent()
	event.EventSource = allowedSources[0]
	payload, err := json.Marshal(event)
	must(err)

	headers := map[string]string{"x-hash": "random", "x-goblin": token}
	response = makeRequest(trackEventURL, headers, payload)

	assert.Equal(t, response["status"], "ERROR")
	assert.Equal(t, response["message"], "Computed hash did not match")
}

func TestKiwiBadEvent(t *testing.T) {
	tokenRequest, err := json.Marshal(getTokenRequest())
	must(err)

	response := makeRequest(getTokenURL, nil, tokenRequest)
	assert.Equal(t, response["status"], "OK")

	token := response["token"].(string)
	assert.NotEmpty(t, token)

	sessionID := uuid.New()
	event := map[string]string{"session_id": sessionID.String(), "event_source": allowedSources[0]}
	payload, err := json.Marshal(event)
	must(err)

	headers := map[string]string{"x-hash": calculateSignature(sessionID, payload), "x-goblin": token}
	response = makeRequest(trackEventURL, headers, payload)

	assert.Equal(t, response["status"], "ERROR")
	assert.Equal(t, response["message"], "Event is malformed")
}

func TestKiwiNoToken(t *testing.T) {
	event := mobileEvent()
	event.EventSource = allowedSources[0]
	payload, err := json.Marshal(event)
	must(err)

	headers := map[string]string{"x-hash": calculateSignature(event.SessionID, payload)}
	response := makeRequest(trackEventURL, headers, payload)

	assert.Equal(t, response["status"], "ERROR")
	assert.Equal(t, response["message"], "Orc token is invalid")
}

func TestKiwiGoodEvent(t *testing.T) {
	tokenRequest, err := json.Marshal(getTokenRequest())
	must(err)

	response := makeRequest(getTokenURL, nil, tokenRequest)
	assert.Equal(t, response["status"], "OK")

	tokenBinary := response["token"].(string)
	assert.NotEmpty(t, tokenBinary)
	token, err := token.Decode(tokenBinary, tokenSecret)
	must(err)

	event := mobileEvent()
	event.EventSource = allowedSources[0]
	payload, err := json.Marshal(event)
	must(err)

	headers := map[string]string{"x-hash": calculateSignature(event.SessionID, payload), "x-goblin": tokenBinary}
	response = makeRequest(trackEventURL, headers, payload)

	assert.Equal(t, response["status"], "OK")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	message, err := Kafka.ReadMessage(ctx)
	must(err)

	var m map[string]interface{}
	json.Unmarshal(message.Value, &m)
	assert.Equal(t, m["event_name"], event.EventName)

	// Should take session&device ids from token
	assert.Equal(t, m["session_id"], token.SessionID.String())
	assert.Equal(t, m["device_info"].(map[string]interface{})["device_id"], token.DeviceID.String())
}

// Orcrist tests

func TestGetTokenNewDevice(t *testing.T) {
    // Get fresh token
	tokenRequest := getTokenRequest()
	payload, err := json.Marshal(tokenRequest)
	must(err)

	response := makeRequest(getTokenURL, nil, payload)
	assert.Equal(t, response["status"], "OK")

	tokenBinary := response["token"].(string)
	assert.NotEmpty(t, tokenBinary)
	token, err := token.Decode(tokenBinary, tokenSecret)
	must(err)

	q := repo.New(DB)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

    // Orcrist should record new device
	device, err := q.GetDevice(ctx, toPgUUID(token.DeviceID))
	must(err)
	assert.Equal(t, tokenRequest.DeviceInfo.DeviceID, toUUID(device.ExternalDeviceID))

    // Orcrist should open new session
    session, err := q.GetSession(ctx, toPgUUID(token.SessionID))
    must(err)
    assert.Equal(t, device.DeviceID, session.DeviceID)
    assert.WithinDuration(t, session.InsertedAt.Time, time.Now(), time.Second * 5)
}

func TestGetTokenExistingDevice(t *testing.T) {
	tokenRequest := getTokenRequest()
	payload, err := json.Marshal(tokenRequest)
	must(err)

	// Record device in Orcrist
	response := makeRequest(getTokenURL, nil, payload)
	assert.Equal(t, response["status"], "OK")

	tokenBinary := response["token"].(string)
	assert.NotEmpty(t, tokenBinary)
	t1, err := token.Decode(tokenBinary, tokenSecret)
	must(err)

	// Artificially expire received token so Orcrist should make new one
	t1.ExpireAt -= 100

	tokenBinary, err = t1.Encode(tokenSecret)
	must(err)

	// Request second token
	response = makeRequest(getTokenURL, map[string]string{"x-goblin": tokenBinary}, payload)
	assert.Equal(t, response["status"], "OK")

	tokenBinary = response["token"].(string)
	assert.NotEmpty(t, tokenBinary)
	t2, err := token.Decode(tokenBinary, tokenSecret)
	must(err)

	// Orcrist should return new token for the same device & session
	assert.Equal(t, t1.DeviceID, t2.DeviceID)
	assert.Equal(t, t1.SessionID, t2.SessionID)
	assert.NotEqual(t, t1, t2)
}

func TestGetTokenFreshToken(t *testing.T) {
	tokenRequest := getTokenRequest()
	payload, err := json.Marshal(tokenRequest)
	must(err)

    // Get fresh token
	response := makeRequest(getTokenURL, nil, payload)
	assert.Equal(t, response["status"], "OK")
	tokenBinary1 := response["token"].(string)
	assert.NotEmpty(t, tokenBinary1)

    // Try to request new token with the same token
    response = makeRequest(getTokenURL, map[string]string{"x-goblin": tokenBinary1}, payload)
	assert.Equal(t, response["status"], "OK")
	tokenBinary2 := response["token"].(string)
	assert.NotEmpty(t, tokenBinary2)

    // Orcrist should return the same token
    assert.Equal(t, tokenBinary1, tokenBinary2)
}

func TestGetTokenInvalidRequest(t *testing.T) {
    payload, err := json.Marshal(map[string]string{"field1": "value1", "field2": "value2"})
    must(err)
	response := makeRequest(getTokenURL, nil, payload)
	assert.Equal(t, response["status"], "ERROR")
	assert.Equal(t, response["message"], "Request is malformed")
}
