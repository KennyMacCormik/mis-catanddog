package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"mis-catanddog/repos"
	"time"
)

// New initializes DB connection
func (s *SqLiteDB) New(uri string, timeout time.Duration) error {
	var err error

	s.db, err = sql.Open("sqlite3", uri)
	if err != nil {
		return fmt.Errorf("failed to create db object: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping repos: %w", err)
	}

	return nil
}

// Get runs SELECT queries
func (s *SqLiteDB) Get(ctx context.Context, r repos.DbReq) (*sql.Rows, error) {
	result, err := s.db.QueryContext(ctx, r.Query, r.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	return result, nil
}

// Exec runs query in transaction and does not return any result
func (s *SqLiteDB) Exec(ctx context.Context, rs []repos.DbReq) error {
	s.m.Lock()
	defer s.m.Unlock()

	// begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to init transaction: %w", err)
	}

	// run all queries inside tx
	for _, val := range rs {
		// prepare statement ot exec
		stmt, err := tx.Prepare(val.Query)
		if err != nil {
			return fmt.Errorf("failed to prepare query %s: %w", val.Query, err)
		}

		// exec statement
		if _, err := stmt.Exec(val.Args...); err != nil {
			return fmt.Errorf("failed query [%s]. Rolling back: %w", val.Query, err)
		}
		defer stmt.Close()
	}

	// commit a transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}
	return nil
}

// Close closes DB connection
func (s *SqLiteDB) Close() {
	s.db.Close()
}
