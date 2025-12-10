package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

var coreTables = []string{
	"users",
	"sessions",
	"messages",
	"langchain_executions",
}

// RunMigrations executes any *.up.sql files in migrationsDir that haven't been applied yet.
// It records applied filenames in schema_migrations to avoid re-running them.
func RunMigrations(ctx context.Context, db *sqlx.DB, migrationsDir string) error {
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		files = append(files, filepath.Join(migrationsDir, name))
	}

	sort.Strings(files)

	missing, err := missingCoreTables(ctx, db)
	if err != nil {
		return fmt.Errorf("check existing tables: %w", err)
	}
	forceApply := len(missing) > 0

	for _, path := range files {
		name := filepath.Base(path)
		if !forceApply {
			applied, err := isApplied(ctx, db, name)
			if err != nil {
				return fmt.Errorf("check migration %s: %w", name, err)
			}
			if applied {
				continue
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}

		statements := splitStatements(string(content))
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("execute migration %s: %w", name, err)
			}
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func isApplied(ctx context.Context, db *sqlx.DB, name string) (bool, error) {
	var exists bool
	if err := db.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, name); err != nil {
		return false, err
	}
	return exists, nil
}

func splitStatements(sql string) []string {
	raw := strings.Split(sql, ";")
	statements := make([]string, 0, len(raw))
	for _, stmt := range raw {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	return statements
}

func missingCoreTables(ctx context.Context, db *sqlx.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		existing[strings.ToLower(name)] = struct{}{}
	}

	var missing []string
	for _, name := range coreTables {
		if _, ok := existing[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing, rows.Err()
}
