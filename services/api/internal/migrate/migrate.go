package migrate

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Up(databaseURL, migrationsDir string) error {
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations path abs: %w", err)
	}
	url := "file://" + filepath.ToSlash(abs)
	m, err := migrate.New(url, databaseURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
