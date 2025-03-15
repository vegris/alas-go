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

	"github.com/google/uuid"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/vegris/alas-go/kiwi/config"
)

// Token represents the decoded token structure.
type Token struct {
	SessionID uuid.UUID `json:"session_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	ExpireAt  int64     `json:"expire_at"`
}

const schemaName = "token.json"
//go:embed token.json
var schemaFS embed.FS
var schema *jsonschema.Schema

func Init() error {
	// Read the embedded schema file
	schemaFile, err := schemaFS.Open(schemaName)
	if err != nil {
		return fmt.Errorf("Failed to read embedded schema: %v", err)
	}

    tokenSchema, err := jsonschema.UnmarshalJSON(schemaFile)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal embedded schema: %v", err)
	}

	// Compile the schema
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaName, tokenSchema); err != nil {
		return fmt.Errorf("Failed to add schema to compiler: %v", err)
	}

	schema, err = compiler.Compile(schemaName)
	if err != nil {
		return fmt.Errorf("Failed to compile schema: %v", err)
	}

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
	block, err := aes.NewCipher(config.Config.TokenSecret)
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
	block, err := aes.NewCipher(config.Config.TokenSecret)
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
	if err := schema.Validate(tokenInstance); err != nil {
		return nil, fmt.Errorf("invalid token JSON: %w", err)
	}

	// Parse JSON into a Token struct
	var token Token
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

