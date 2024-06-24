package repos

import (
	"context"
	"database/sql"
	"time"
)

type DbReq struct {
	Query string
	Args  []any
}

type DB interface {
	New(uri string, timeout time.Duration) error
	Get(ctx context.Context, r DbReq) (*sql.Rows, error)
	Exec(ctx context.Context, rs []DbReq) error
	Close()
}
