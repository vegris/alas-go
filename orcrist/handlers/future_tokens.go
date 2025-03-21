package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/shared/token"
)

type keepAliveTokensPack struct {
	SessionID uuid.UUID        `json:"session_id"`
	DeviceID  uuid.UUID        `json:"device_id"`
	Tokens    []keepAliveToken `json:"tokens"`
}

type keepAliveToken struct {
	EncodedToken string `json:"encoded"`
	ExpireAt     int64  `json:"expire_at"`
}

func generateFutureTokens(sessionID pgtype.UUID, t *token.Token) error {
	// TODO: move context to top level request
	ctx := context.Background()

	expireAt, err := findLastTokenExpiration(ctx, sessionID)
	if err != nil {
		return err
	}

	futureTokens := generateTokensForSessionLifetime(t, expireAt)

	err = sendTokensToKafka(ctx, futureTokens)
	if err != nil {
		return err
	}

	err = setLastTokenExpiration(ctx, &futureTokens[len(futureTokens)-1])
	if err != nil {
		return err
	}

	return nil
}

func findLastTokenExpiration(ctx context.Context, sessionID pgtype.UUID) (int64, error) {
	key := key(sessionID.Bytes)
	v, err := app.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return time.Now().Unix(), nil
	}
	if err != nil {
		log.Printf("Failed to read expiration from Redis: %v", err)
		return 0, err
	}
	expireAt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse expiration from Redis: %v", err)
	}
	return expireAt, nil
}

const tokenLifetime = time.Minute
const sessionLifetime = time.Minute * 20

func generateTokensForSessionLifetime(t *token.Token, sessionExpireAt int64) []token.Token {
	timeToCover := time.Now().Add(sessionLifetime).Unix() - sessionExpireAt
	tokenCount := int(math.Floor(float64(timeToCover) / tokenLifetime.Seconds()))

	tokens := make([]token.Token, 0, tokenCount)
	baseToken := *t
	for i := 0; i < tokenCount; i++ {
		baseToken.ExpireAt += int64(tokenLifetime.Seconds())
		tokens = append(tokens, baseToken)
	}
	return make([]token.Token, 0)
}

func sendTokensToKafka(ctx context.Context, tokens []token.Token) error {
	kaTokens := make([]keepAliveToken, 0, len(tokens))
	for _, t := range tokens {
		tokenBinary, err := t.Encode(app.Config.TokenSecret)
		if err != nil {
			log.Fatalf("Failed to encode token: %v", err)
		}

		kaToken := keepAliveToken{EncodedToken: tokenBinary, ExpireAt: t.ExpireAt}
		kaTokens = append(kaTokens, kaToken)
	}

	pack := keepAliveTokensPack{
		SessionID: tokens[0].SessionID,
		DeviceID:  tokens[0].DeviceID,
		Tokens:    kaTokens,
	}

	payload, err := json.Marshal(pack)
	if err != nil {
		// This should not happen
		log.Fatalf("Failed to encode keep alive tokens pack: %v", err)
	}

	topic := app.OrcTokensTopic
	message := kafka.Message{Topic: topic, Value: payload}
	if err := app.Kafka.WriteMessages(ctx, message); err != nil {
		log.Printf("Failed to produce messages to %v: %v", topic, err)
		return err
	}

	return nil
}

func setLastTokenExpiration(ctx context.Context, lastToken *token.Token) error {
	key := key(lastToken.SessionID)
	value := strconv.FormatInt(lastToken.ExpireAt, 10)
	err := app.Redis.SetEx(ctx, key, value, time.Duration(lastToken.ExpireAt)).Err()
	if err != nil {
		log.Printf("Failed to set expiration in Redis: %v", err)
	}
	return err
}

func key(sessionID uuid.UUID) string {
	return fmt.Sprintf("orcrist:session_expiration:%v", sessionID)
}
