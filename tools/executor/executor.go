package executor

import (
	"context"
	"database/sql"
)

type Executor interface {
	// ExecContext выполняет запрос без возврата строк (INSERT, UPDATE, DELETE)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryContext выполняет запрос с возвратом строк (SELECT)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext выполняет запрос с возвратом одной строки
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
