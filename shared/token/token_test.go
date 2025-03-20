package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var secret []byte

func TestMain(m *testing.M) {
	Initialize()

	s, err := DecodeSecret("vk5LxATdFF6whkWrTIs5UXsQOD1gbjGWIecKSsf4Q5I=")
	if err != nil {
		log.Fatalf("Failed to decode token secret: %v", err)
	}
	secret = s

	code := m.Run()
	os.Exit(code)
}

func TestEncodeDecode(t *testing.T) {
	// Create a sample token
	sessionID := uuid.New()
	deviceID := uuid.New()
	expireAt := time.Now().Add(24 * time.Hour).Unix()

	token := Token{
		SessionID: sessionID,
		DeviceID:  deviceID,
		ExpireAt:  expireAt,
	}

	// Encode the token
	encoded, err := token.Encode(secret)
	assert.NoError(t, err, "Encode should not return an error")
	assert.NotEmpty(t, encoded, "Encoded token should not be empty")

	// Decode the token
	decoded, err := Decode(encoded, secret)
	assert.NoError(t, err, "Decode should not return an error")
	assert.Equal(t, token.SessionID, decoded.SessionID, "SessionID should match")
	assert.Equal(t, token.DeviceID, decoded.DeviceID, "DeviceID should match")
	assert.Equal(t, token.ExpireAt, decoded.ExpireAt, "ExpireAt should match")
}

func TestDecodeInvalidToken(t *testing.T) {
	// Test with an invalid Base64 string
	_, err := Decode("invalid-base64", secret)
	assert.Error(t, err, "Decode should return an error for invalid Base64")

	// Test with a token that's too short
	shortToken := base64.URLEncoding.EncodeToString([]byte("too-short"))
	_, err = Decode(shortToken, secret)
	assert.Error(t, err, "Decode should return an error for a token that's too short")

	// Test with a tampered token (invalid JSON)
	tamperedToken := make([]byte, 32)
	rand.Read(tamperedToken)
	encodedTamperedToken := base64.URLEncoding.EncodeToString(tamperedToken)
	_, err = Decode(encodedTamperedToken, secret)
	assert.Error(t, err, "Decode should return an error for a tampered token")
}

func TestDecodeInvalidJSON(t *testing.T) {
	// Create a valid token but with invalid JSON content
	invalidJSON := []byte(`{"session_id": "invalid", "device_id": "invalid", "expire_at": "invalid"}`)
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	block, err := aes.NewCipher(secret)
	assert.NoError(t, err, "Failed to create cipher")

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(invalidJSON))
	stream.XORKeyStream(ciphertext, invalidJSON)

	combined := append(iv, ciphertext...)
	encodedToken := base64.URLEncoding.EncodeToString(combined)

	// Attempt to decode
	_, err = Decode(encodedToken, secret)
	assert.Error(t, err, "Decode should return an error for invalid JSON")
}
