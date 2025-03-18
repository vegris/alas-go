package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/vegris/alas-go/shared/schemas"
)

//go:embed token.json
var schemaFS embed.FS
var tokenSchema *jsonschema.Schema

// Token represents the decoded token structure.
type Token struct {
	SessionID uuid.UUID `json:"session_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	ExpireAt  int64     `json:"expire_at"`
}

func Initialize() {
	schemaFile, err := schemaFS.Open("token.json")
	if err != nil {
		log.Fatalf("Failed to open schema file: %v", err)
	}

	schema, err := schemas.CompileSchema(schemaFile)
	if err != nil {
		log.Fatalf("Failed to compile token schema: %v", err)
	}

	tokenSchema = schema
}

func DecodeSecret(value string) ([]byte, error) {
	secret, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
        return nil, err
	}
	return secret, nil
}

// Encode encodes a Token into a Base64-encoded encrypted string.
func (token Token) Encode(secret []byte) (string, error) {
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
	block, err := aes.NewCipher(secret)
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
func Decode(encodedToken string, secret []byte) (*Token, error) {
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
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	// Parse into generic interface{} for schema validation
	var tokenInstance any
	if err := json.Unmarshal(plaintext, &tokenInstance); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate JSON against the schema
	if err := tokenSchema.Validate(tokenInstance); err != nil {
		return nil, fmt.Errorf("invalid token JSON: %w", err)
	}

	// Parse JSON into a Token struct
	var token Token
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}
