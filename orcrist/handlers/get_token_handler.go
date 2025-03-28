package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/events"
	"github.com/vegris/alas-go/orcrist/sessions"
	"github.com/vegris/alas-go/shared/token"
)

type tokenRequestResponse struct {
	Status   string `json:"status"`
	Token    string `json:"token"`
	TokenTTL int64  `json:"ttl"`
}

type errResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	errBadToken     = errors.New("Orc token is invalid")
	errReadError    = errors.New("Failed to read request")
	errRequestError = errors.New("Request is malformed")
)

func HandleGetToken(w http.ResponseWriter, r *http.Request) {
	tokenBinary := r.Header.Get("x-goblin")

	oldToken, err := getOldToken(tokenBinary)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return the same token if it's fresh enough
	if oldToken != nil && !isTokenExpired(oldToken) {
		json.NewEncoder(w).Encode(tokenRequestResponse{Status: "OK", Token: tokenBinary, TokenTTL: oldToken.ExpireAt})
		return
	}

	body, err := readBody(r)
	if err != nil {
		handleError(w, err)
		return
	}

	request, err := parseRequest(body)
	if err != nil {
		handleError(w, err)
		return
	}

	var token *token.Token
	if oldToken == nil {
		token = sessions.CreateToken(request)
	} else {
		token = sessions.RefreshToken(request, oldToken)
	}

	err = generateFutureTokens(request.SessionID, token)
	if err != nil {
		// Do not halt request processing on this
		// It's still possible to return fresh token to the user
		log.Printf("Error generating future tokens: %v", tokenBinary)
	}

	tokenBinary, err = token.Encode(app.Config.TokenSecret)
	if err != nil {
		log.Fatalf("Error encoding token: %v", tokenBinary)
	}

	json.NewEncoder(w).Encode(tokenRequestResponse{Status: "OK", Token: tokenBinary, TokenTTL: token.ExpireAt})
}

func getOldToken(tokenBinary string) (*token.Token, error) {
	if tokenBinary == "" {
		return nil, nil
	}

	t, err := token.Decode(tokenBinary, app.Config.TokenSecret)
	if err != nil {
		return nil, errBadToken
	}

	return t, nil
}

func isTokenExpired(token *token.Token) bool {
	return token.ExpireAt <= time.Now().Unix()
}

func readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errReadError
	}
	defer r.Body.Close()
	return body, nil
}

func parseRequest(body []byte) (*events.GetTokenRequest, error) {
	request, err := events.ParseGetTokenRequest(body)
	if err != nil {
		return nil, errRequestError
	}
	return request, nil
}

func handleError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(errResponse{
		Status:  "ERROR",
		Message: err.Error(),
	})
}
