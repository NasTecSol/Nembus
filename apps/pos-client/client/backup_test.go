package client

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type recordingExecutor struct {
	statements []string
	failAt     int
	err        error
}

func (e *recordingExecutor) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	e.statements = append(e.statements, sql)
	if len(e.statements) == e.failAt {
		return pgconn.CommandTag{}, e.err
	}
	return pgconn.CommandTag{}, nil
}

func TestExecuteStatementsStopsOnFailure(t *testing.T) {
	for _, code := range []string{"53100", "23505", "42P01", "42P07"} {
		t.Run(code, func(t *testing.T) {
			failure := &pgconn.PgError{Code: code, Message: "restore failed"}
			exec := &recordingExecutor{failAt: 2, err: failure}
			err := executeStatements(context.Background(), exec, "SELECT 1; SELECT 2; SELECT 3;")
			if !errors.Is(err, failure) {
				t.Fatalf("expected original PostgreSQL failure, got %v", err)
			}
			if len(exec.statements) != 2 {
				t.Fatalf("restore continued after failure: %v", exec.statements)
			}
		})
	}
}

func TestExecuteStatementsAllowsExistingSchemaAndSkipsCloudOwner(t *testing.T) {
	exec := &recordingExecutor{failAt: 1, err: &pgconn.PgError{Code: "42P06"}}
	err := executeStatements(context.Background(), exec,
		"CREATE SCHEMA public; ALTER SCHEMA public OWNER TO cloud_user; SELECT 1;")
	if err != nil {
		t.Fatal(err)
	}
	if len(exec.statements) != 2 || exec.statements[1] != "SELECT 1" {
		t.Fatalf("unexpected restored statements: %v", exec.statements)
	}
}
