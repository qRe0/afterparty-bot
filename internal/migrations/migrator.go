package migrator

import (
	"embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var (
	//go:embed *.sql
	embedMigrations embed.FS
)

const (
	migrationDir = "."
)

type Migration struct {
	Database *sqlx.DB
}

type Migrator interface {
	Up() error
	Down() error
	Latest() error
}

func New(db *sqlx.DB) (Migrator, error) {
	goose.SetBaseFS(embedMigrations)

	err := goose.SetDialect("postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	return &Migration{Database: db}, nil
}

func (m *Migration) Up() error {
	err := goose.Up(m.Database.DB, migrationDir)
	if err != nil {
		return fmt.Errorf("failed to migrate db UP: %w", err)
	}

	return nil
}

func (m *Migration) Down() error {
	err := goose.Down(m.Database.DB, migrationDir)
	if err != nil {
		return fmt.Errorf("failed to migrate db DOWN: %w", err)
	}

	return nil
}

func (m *Migration) Latest() error {
	latest, err := getLatestMigrationVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest migration version: %w", err)
	}

	err = goose.UpTo(m.Database.DB, migrationDir, latest)
	if err != nil {
		return fmt.Errorf("failed to migrate db to latest version: %w", err)
	}

	return nil
}

func getLatestMigrationVersion() (int64, error) {
	files, err := os.ReadDir("./internal/migrations")
	if err != nil {
		return 0, err
	}

	var latestVersion int64
	rx := regexp.MustCompile(`^(\d+)_.*\.sql$`)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		matches := rx.FindStringSubmatch(file.Name())

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return 0, err
		}

		if version > latestVersion {
			latestVersion = version
		}
	}

	return latestVersion, nil
}
