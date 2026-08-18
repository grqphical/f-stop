package database

import (
	"embed"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func applyMigrations(databaseURL string) {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("migration source error: %v\n", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		log.Fatalf("migration instance error: %v\n", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration error: %v\n", err)
	}
}
