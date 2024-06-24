package sqlite3

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type SqLiteDB struct {
	db *sql.DB
	m  sync.Mutex // sqlite poorly handles simultaneous writes
}
