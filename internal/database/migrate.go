package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
)

// applyMigrations applies all up migrations from the migrations directory
func ApplyMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations", // Path to your migration files
		databaseURL,
	)
	if err != nil {
		return err
	}

	// Run the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully.")
	return nil
}
