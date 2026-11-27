package worker

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // the "pgx" database/sql driver goose migrates through
	"github.com/pressly/goose/v3"
)

// TestMain gives the worker's tests a database of their own, migrated fresh.
// `go test ./...` runs packages in parallel, and the worker's check functions
// take every due monitor in the database (ListDue*Monitors): on the shared one
// they checked the handler tests' monitors too, which then counted a check
// they never seeded and failed at random.
func TestMain(m *testing.M) {
	if err := useOwnDatabase("checkmeup_worker_test"); err != nil {
		fmt.Fprintln(os.Stderr, "worker tests: own database:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// useOwnDatabase recreates name on the test database's server, migrates it,
// and points DATABASE_URL, which testPool reads, at it.
func useOwnDatabase(name string) error {
	base := testDatabaseURL()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, base)
	if err != nil {
		return fmt.Errorf("connect to test db: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
		return fmt.Errorf("drop %s: %w", name, err)
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}

	u, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	u.Path = "/" + name
	own := u.String()

	sqlDB, err := sql.Open("pgx", own)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer func() { _ = sqlDB.Close() }()
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(sqlDB, "../../migrations"); err != nil {
		return fmt.Errorf("migrate %s: %w", name, err)
	}
	return os.Setenv("DATABASE_URL", own)
}
