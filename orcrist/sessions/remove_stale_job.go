package sessions

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vegris/alas-go/orcrist/app"
	"github.com/vegris/alas-go/orcrist/repo"
)

func RemoveStaleSessions(ctx context.Context) {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	timer := time.NewTimer(midnight.Sub(now))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		runJob(ctx, midnight)
	}
}

const jobName = "remove_stale_sessions"
const staleSessionDays = 3

func runJob(ctx context.Context, scheduledTime time.Time) {
	timestamp := pgtype.Timestamp{Time: scheduledTime, Valid: true}

	tx, err := app.DB.Begin(ctx)
	if err != nil {
		log.Printf("Failed to open transaction: %v", err)
		return
	}
	defer tx.Rollback(ctx)
	q := repo.New(app.DB).WithTx(tx)

	jobParams := repo.InsertJobParams{Name: jobName, RunAt: timestamp}
	if _, err := q.InsertJob(ctx, jobParams); err != nil {
		log.Printf("Failed to insert job: %v", err)
		return
	}

	if err := q.RemoveStaleSessions(ctx, staleSessionDays); err != nil {
		log.Printf("Failed to remove stale sessions: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	q = repo.New(app.DB)
	params := repo.RemoveOldJobsParams{Name: jobName, RunAt: timestamp}
	if err := q.RemoveOldJobs(ctx, params); err != nil {
		log.Printf("Failed to clear previous jobs: %v", err)
	}

	log.Println("Successfully removed stale sessions!")
}
