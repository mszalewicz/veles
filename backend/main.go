package backend

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Ensure that schema used inside database, matches schema used during bindings created via sqlc
//
//go:embed sql/schema.sql
var schema string

var Q *Queries
var DB *sql.DB
var mu sync.Mutex

// Connecting to SQLite
func Connect(dbPath string) error {
	var err error

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("Failed to open SQLite: %w", err)
	}

	_, err = DB.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;
	`)
	if err != nil {
		return fmt.Errorf("Failed to set PRAGMAs: %w", err)
	}

	// Initialize the sqlc-generated queries
	Q = New(DB)

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func ApplySchema() error {
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("Error applying schema: %w", err)
	}

	return nil
}

// Write (For queries with no arguments that return a value)
func Write[R any](ctx context.Context, fn func(*Queries, context.Context) (R, error)) (R, error) {
	mu.Lock()
	defer mu.Unlock()

	var res R
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// database/sql uses a connection pool, but SQLite supports only one writer at a time. To prevent SQLITE_BUSY errors, we use WithTx and a mutex to:
	// 1) ensure atomicity: provide automatic rollbacks on failure
	// 2) serialize access: pin writes to a single connection and prevent concurrent write attempts from multiple goroutines
	res, err = fn(Q.WithTx(tx), ctx)
	if err != nil {
		return res, err
	}

	return res, tx.Commit()
}

// WriteArg (for queries with arguments that return a value)
// Note: sqlc usually bundles multiple arguments into a single Params struct.
func WriteArg[A any, R any](ctx context.Context, fn func(*Queries, context.Context, A) (R, error), arg A) (R, error) {
	mu.Lock()
	defer mu.Unlock()

	var res R
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// database/sql uses a connection pool, but SQLite supports only one writer at a time. To prevent SQLITE_BUSY errors, we use WithTx and a mutex to:
	// 1) ensure atomicity: provide automatic rollbacks on failure
	// 2) serialize access: pin writes to a single connection and prevent concurrent write attempts from multiple goroutines
	res, err = fn(Q.WithTx(tx), ctx, arg)
	if err != nil {
		return res, err
	}

	return res, tx.Commit()
}

// ExecArg (for queries with arguments that only return an error, like Updates/Deletes)
func ExecArg[A any](ctx context.Context, arg A, fn func(*Queries, context.Context, A) error) error {
	mu.Lock()
	defer mu.Unlock()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// database/sql uses a connection pool, but SQLite supports only one writer at a time. To prevent SQLITE_BUSY errors, we use WithTx and a mutex to:
	// 1) ensure atomicity: provide automatic rollbacks on failure
	// 2) serialize access: pin writes to a single connection and prevent concurrent write attempts from multiple goroutines
	err = fn(Q.WithTx(tx), ctx, arg)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Lock safe queries:

func SafeInsertDefaultWindow(ctx context.Context, arg InsertDefaultWindowParams) error {
	return ExecArg(context.Background(), arg, func(q *Queries, ctx context.Context, arg InsertDefaultWindowParams) error {
		err := q.InsertDefaultWindow(ctx, arg)
		return err
	})
}

func SafeUpdateWindowGeometry(ctx context.Context, arg UpdateWindowGeometryParams) error {
	return ExecArg(context.Background(), arg, func(q *Queries, ctx context.Context, arg UpdateWindowGeometryParams) error {
		err := q.UpdateWindowGeometry(ctx, arg)
		return err
	})
}

func SafeSqliteUserVersion(user_version int) error {
        tx, err := DB.Begin()
        if err != nil {
            return err
        }

        defer tx.Rollback()

        query := fmt.Sprintf("PRAGMA user_version = %d", user_version)

        _, err = tx.Exec(query)
        if err != nil {
            return err
        }

        return tx.Commit()
}