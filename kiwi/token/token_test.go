package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vegris/alas-go/kiwi/config"
	"github.com/vegris/alas-go/kiwi/schemas"
)

func TestMain(m *testing.M) {
    config.Initialize()
    schemas.Initialize()
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
	encoded, err := token.Encode()
	assert.NoError(t, err, "Encode should not return an error")
	assert.NotEmpty(t, encoded, "Encoded token should not be empty")

	// Decode the token
	decoded, err := Decode(encoded)
	assert.NoError(t, err, "Decode should not return an error")
	assert.Equal(t, token.SessionID, decoded.SessionID, "SessionID should match")
	assert.Equal(t, token.DeviceID, decoded.DeviceID, "DeviceID should match")
	assert.Equal(t, token.ExpireAt, decoded.ExpireAt, "ExpireAt should match")
}

func TestDecodeInvalidToken(t *testing.T) {
	// Test with an invalid Base64 string
	_, err := Decode("invalid-base64")
	assert.Error(t, err, "Decode should return an error for invalid Base64")

	// Test with a token that's too short
	shortToken := base64.URLEncoding.EncodeToString([]byte("too-short"))
	_, err = Decode(shortToken)
	assert.Error(t, err, "Decode should return an error for a token that's too short")

	// Test with a tampered token (invalid JSON)
	tamperedToken := make([]byte, 32)
	rand.Read(tamperedToken)
	encodedTamperedToken := base64.URLEncoding.EncodeToString(tamperedToken)
	_, err = Decode(encodedTamperedToken)
	assert.Error(t, err, "Decode should return an error for a tampered token")
}

func TestDecodeInvalidJSON(t *testing.T) {
	// Create a valid token but with invalid JSON content
	invalidJSON := []byte(`{"session_id": "invalid", "device_id": "invalid", "expire_at": "invalid"}`)
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	block, err := aes.NewCipher(config.Config.TokenSecret)
	assert.NoError(t, err, "Failed to create cipher")

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(invalidJSON))
	stream.XORKeyStream(ciphertext, invalidJSON)

	combined := append(iv, ciphertext...)
	encodedToken := base64.URLEncoding.EncodeToString(combined)

	// Attempt to decode
	_, err = Decode(encodedToken)
	assert.Error(t, err, "Decode should return an error for invalid JSON")
}
