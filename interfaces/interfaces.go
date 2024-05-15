package interfaces

import (
	"context"
	"database/sql"
	"time"
)

type DB interface {
	New(uri, dbType string) error
	Get(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, q string, args ...any) error
	Init(timeout time.Duration) error
	Close()
}
