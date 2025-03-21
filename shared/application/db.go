package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StartPostgres(dbName string, migrations fs.FS) *pgxpool.Pool {
	const postgresURL = "postgres://postgres:postgres@localhost:5432"

	ctx := context.Background()

	db, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		log.Fatalf("Unable to connect to Postgres: %v", err)
	}

	// Create database is not exists
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err := db.Exec(ctx, query); err == nil {
		log.Println("Application database created successfully!")
	} else {
		// Continue starting Postgres on create db error
		// because the db can already exist
		log.Printf("Failed to create application database: %v", err)
	}

	if err := db.Close(ctx); err != nil {
		log.Fatalf("Failed to close DB connection: %v", err)
	}

	// Run migrations
	sourceDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatalf("Failed to create source driver for migrations: %v", err)
	}

	dbURL := "pgx5://postgres:postgres@localhost:5432/" + dbName
	migrator, err := migrate.NewWithSourceInstance("migrator", sourceDriver, dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	log.Println("Successfully applied DB migrations!")

	sourceErr, dbErr := migrator.Close()
	if sourceErr != nil {
		log.Fatalf("Failed to close source migration driver: %v", err)
	}
	if dbErr != nil {
		log.Fatalf("Failed to close database migration driver: %v", err)
	}

	// Initialize connection pool
	dbpool, err := pgxpool.New(ctx, fmt.Sprintf("%s/%s", postgresURL, dbName))
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DB via connection pool: %v", err)
	}

	log.Println("Postgres connection pool initialized!")

	return dbpool
}

func ShutdownPostgres(dbpool *pgxpool.Pool) {
	dbpool.Close()
	log.Println("Postgres connection pool successfully closed!")
}
