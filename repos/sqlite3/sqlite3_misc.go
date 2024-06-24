package sqlite3

import (
	"context"
	"fmt"
	"time"
)

// Init tries to create necessary tables if they don't exist
func (s *SqLiteDB) Init(timeout time.Duration) error {
	if s.db == nil {
		return fmt.Errorf("db connection is not set")
	}

	var q string = "CREATE TABLE IF NOT EXISTS `doc_type` ( \t`id` integer primary key NOT NULL UNIQUE, \t`doc` TEXT NOT NULL ); CREATE TABLE IF NOT EXISTS `animal_type` ( \t`id` integer primary key NOT NULL UNIQUE, \t`type` TEXT NOT NULL ); CREATE TABLE IF NOT EXISTS `human` ( \t`doc_id` integer primary key NOT NULL UNIQUE, \t`doc_type` INTEGER NOT NULL, \t`first_name` TEXT NOT NULL, \t`middle_name` TEXT, \t`last_name` TEXT NOT NULL, \t`birth_date` REAL NOT NULL, FOREIGN KEY(`doc_type`) REFERENCES `doc_type`(`id`) ); CREATE TABLE IF NOT EXISTS `animal` ( \t`doc_id` integer primary key NOT NULL UNIQUE, \t`doc_type` INTEGER NOT NULL, \t`name` TEXT NOT NULL, \t`birth_date` REAL NOT NULL, \t`animal_type` INTEGER NOT NULL, \t`breed` TEXT NOT NULL, \t`owner_doc_id` TEXT NOT NULL, FOREIGN KEY(`doc_type`) REFERENCES `doc_type`(`id`), FOREIGN KEY(`animal_type`) REFERENCES `animal_type`(`id`), FOREIGN KEY(`owner_doc_id`) REFERENCES `human`(`doc_id`) );"

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("init db error: %w", err)
	}

	return nil
}

// ForceInitDictTables deletes everything dictionary tables and writes them with data
func (s *SqLiteDB) ForceInitDictTables(timeout time.Duration) error {
	var qList = []string{
		"DELETE FROM doc_type;",
		"DELETE FROM animal_type;",
		"INSERT INTO doc_type (id,doc) VALUES (1, 'passport'),(2, 'veterinary passport'),(3, 'military passport');",
		"INSERT INTO animal_type (id,type) VALUES (1, 'dog'),(2, 'cat');",
	}

	if s.db == nil {
		return fmt.Errorf("db connection is not set")
	}

	for _, val := range qList {
		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
		defer cancel()

		if _, err := s.db.ExecContext(ctx, val); err != nil {
			return fmt.Errorf("failed to init dicts: %w", err)
		}
	}

	return nil
}
