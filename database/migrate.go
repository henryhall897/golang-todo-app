package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // Import PostgreSQL driver
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// ApplyMigrations applies all database migrations using embedded migration files.
func ApplyMigrations(ctx context.Context, databaseURL string) error {
	log.Println("Starting database migrations...")

	// Open the database connection
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Initialize migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Use embedded filesystem for migrations
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply migrations
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply.")
			return nil
		}
		log.Printf("Migration failed: %v. Attempting rollback...", err)

		// Attempt rollback
		if rollbackErr := m.Steps(-1); rollbackErr != nil {
			log.Printf("Rollback failed: %v", rollbackErr)
			return fmt.Errorf("migration failed: %v; rollback also failed: %w", err, rollbackErr)
		}

		log.Println("Migration rollback successful.")
		return fmt.Errorf("migration failed and rolled back successfully: %w", err)
	}

	log.Println("Database migrations applied successfully.")
	return nil
}
