package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

type SqLiteDB struct {
	db *sql.DB
	m  sync.Mutex // sqlite poorly handles simultaneous writes
}

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
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Init tries to create necessary tables if they don't exist
func (s *SqLiteDB) Init(timeout time.Duration) error {
	if s.db == nil {
		return fmt.Errorf("db connection is not set")
	}

	var q string = "CREATE TABLE IF NOT EXISTS `doc_type` ( \t`id` integer primary key NOT NULL UNIQUE, \t`doc` TEXT NOT NULL ); CREATE TABLE IF NOT EXISTS `animal_type` ( \t`id` integer primary key NOT NULL UNIQUE, \t`type` TEXT NOT NULL ); CREATE TABLE IF NOT EXISTS `human` ( \t`doc_id` integer primary key NOT NULL UNIQUE, \t`doc_type` INTEGER NOT NULL, \t`firsst_name` TEXT NOT NULL, \t`middle_name` TEXT, \t`last_name` TEXT NOT NULL, \t`birth_date` REAL NOT NULL, FOREIGN KEY(`doc_type`) REFERENCES `doc_type`(`id`) ); CREATE TABLE IF NOT EXISTS `animal` ( \t`doc_id` integer primary key NOT NULL UNIQUE, \t`doc_type` INTEGER NOT NULL, \t`name` TEXT NOT NULL, \t`birth_date` REAL NOT NULL, \t`animal_type` INTEGER NOT NULL, \t`breed` TEXT NOT NULL, \t`owner_doc_id` TEXT NOT NULL, FOREIGN KEY(`doc_type`) REFERENCES `doc_type`(`id`), FOREIGN KEY(`animal_type`) REFERENCES `animal_type`(`id`), FOREIGN KEY(`owner_doc_id`) REFERENCES `human`(`doc_id`) );"

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("init db error: %w", err)
	}

	return nil
}

// Get runs SELECT queries
func (s *SqLiteDB) Get(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	if q[:6] != "SELECT" {
		return nil, fmt.Errorf("query %s is not a SELECT query", q)
	}
	result, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	return result, nil
}

// Exec runs query in transaction and does not return any result
func (s *SqLiteDB) Exec(ctx context.Context, q string, args ...any) error {
	s.m.Lock()
	defer s.m.Unlock()

	// begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to init transaction: %w", err)
	}

	// prepare statement ot exec
	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare query %s: %w", q, err)
	}
	defer stmt.Close()

	// exec statement
	if _, err := stmt.Exec(args...); err != nil {
		return fmt.Errorf("failed query [%s]. Rolling back: %w", q, err)
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
