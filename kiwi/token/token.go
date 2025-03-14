package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// Token represents the decoded token structure.
type Token struct {
	SessionID uuid.UUID `json:"session_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	ExpireAt  int64     `json:"expire_at"`
}

// secretKey is the AES256 secret key used for encryption/decryption.
var secretKey []byte

// Init initializes the package by loading the secret key from the environment.
func Init() error {
	key := os.Getenv("TOKEN_SECRET")
	if key == "" {
		return errors.New("TOKEN_SECRET environment variable is not set")
	}

    token, err := base64.StdEncoding.DecodeString(key)
    if err != nil {
        return fmt.Errorf("failed to decode TOKEN_SECRET: %w", err)
    }
    secretKey = token
    return nil
}

// Encode encodes a Token into a Base64-encoded encrypted string.
func (token Token) Encode() (string, error) {
	// Marshal the token into JSON
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt the token body using AES256 CTR
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(tokenJSON))
	stream.XORKeyStream(ciphertext, tokenJSON)

	// Combine IV and ciphertext
	combined := append(iv, ciphertext...)

	// Base64 encode the result
	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decode decodes a Base64-encoded encrypted string into a Token.
func Decode(encodedToken string) (*Token, error) {
	// Base64 decode the token
	combined, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Split into IV and ciphertext
	if len(combined) <= aes.BlockSize {
		return nil, errors.New("invalid token: too short")
	}
	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]

	// Decrypt the ciphertext using AES256 CTR
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	// Parse the JSON into a Token struct
	var token Token
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}
