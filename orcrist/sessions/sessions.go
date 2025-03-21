package sessions

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/events"
	"github.com/vegris/alas-go/orcrist/repo"
	"github.com/vegris/alas-go/shared/token"
)

const sessionDuration = time.Minute

var sessionDurationSecs = int64(sessionDuration.Seconds())

func RefreshToken(request *events.GetTokenRequest, t *token.Token) *token.Token {
	ctx := context.Background()

	session := refreshExistingSession(ctx, request.SessionID)
	if session == nil {
		session = createSession(ctx, request)
	}
	return createToken(session)
}

func CreateToken(request *events.GetTokenRequest) *token.Token {
	ctx := context.Background()
	session := createSession(ctx, request)
	return createToken(session)
}

func refreshExistingSession(ctx context.Context, sessionID pgtype.UUID) *repo.Session {
	q := repo.New(app.DB)

	session, err := q.GetAliveSession(ctx, sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		log.Printf("Failed to query repo for session: %v", err)
		return fakeSession()
	}

	params := repo.RefreshSessionParams{SessionID: session.SessionID, SessionDuration: sessionDurationSecs}
	session, err = q.RefreshSession(ctx, params)
	if err != nil {
		log.Printf("Failed to update session: %v", err)
		return fakeSession()
	}

	return &session
}

func createSession(ctx context.Context, request *events.GetTokenRequest) *repo.Session {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return fakeSession()
	}
	defer tx.Rollback(ctx)
	q := repo.New(app.DB).WithTx(tx)

	device, err := getOrCreateDevice(ctx, tx, request)
	if err != nil {
		return fakeSession()
	}

	params := repo.CreateSessionParams{SessionID: genUUID(), DeviceID: device.DeviceID, SessionDuration: sessionDurationSecs}
	session, err := q.CreateSession(ctx, params)
	if err != nil {
		return fakeSession()
	}
	tx.Commit(ctx)

	return &session
}

func getOrCreateDevice(ctx context.Context, tx pgx.Tx, request *events.GetTokenRequest) (device repo.Device, err error) {
	q := repo.New(app.DB).WithTx(tx)

	device, err = q.GetDeviceByExternalDeviceID(ctx, request.DeviceInfo.DeviceID)
	if err == nil {
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("Failed to query repo for device: %v", err)
		return
	}

	params := repo.CreateDeviceParams{
		DeviceID:         genUUID(),
		Source:           request.EventSource,
		ExternalDeviceID: request.DeviceInfo.DeviceID,
		// TODO: pass metadata here
	}
	device, err = q.CreateDevice(ctx, params)
	if err != nil {
		log.Printf("Failed to create device: %v", err)
	}
	return
}

func createToken(session *repo.Session) *token.Token {
	return &token.Token{
		SessionID: session.SessionID.Bytes,
		DeviceID:  session.DeviceID.Bytes,
		ExpireAt:  session.EndsAt.Time.Unix(),
	}
}

func fakeSession() *repo.Session {
	now := time.Now()

	return &repo.Session{
		SessionID:  genUUID(),
		DeviceID:   genUUID(),
		InsertedAt: toTimestamp(now),
		UpdatedAt:  toTimestamp(now),
		EndsAt:     toTimestamp(now.Add(sessionDuration)),
	}
}

func genUUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	}
}

func toTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t}
}
