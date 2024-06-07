package database

import (
	"context"
	"os"
	"testing"
	"time"
)

var db = &SqLiteDB{}

const dbTimeout int = 1000
const testDbUri string = "test.sqlite3"

func TestSqLiteDB_New(t *testing.T) {
	err := db.New(testDbUri, time.Duration(dbTimeout)*time.Millisecond)
	if err != nil {
		t.Fatalf("Can not create DB: %s", err.Error())
	}
}

func TestSqLiteDB_Init(t *testing.T) {
	if err := db.Init(time.Duration(dbTimeout)); err != nil {
		t.Fatalf("Failed to init DB: %s", err.Error())
	}
}

func TestSqLiteDB_Init_Negative(t *testing.T) {
	//s.db == nil
	db := &SqLiteDB{}
	if err := db.Init(time.Duration(dbTimeout)); err.Error() != "db connection is not set" {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestSqLiteDB_ForceInitDictTables(t *testing.T) {
	if err := db.ForceInitDictTables(time.Duration(dbTimeout)); err != nil {
		t.Fatalf("failed to init dicts: %s", err.Error())
	}
}

func TestSqLiteDB_ForceInitDictTables_Negative(t *testing.T) {
	//s.db == nil
	db := &SqLiteDB{}
	if err := db.ForceInitDictTables(time.Duration(dbTimeout)); err.Error() != "db connection is not set" {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestSqLiteDB_Exec(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbTimeout)*time.Millisecond)
	defer cancel()

	if err := db.Exec(ctx, "INSERT INTO doc_type (doc) VALUES (?)", "Test Passport"); err != nil {
		t.Fatalf("Can not insert values to DB: %s", err.Error())
	}
}

func TestSqLiteDB_Get(t *testing.T) {
	type testCase struct {
		query     string
		valideVal string
		err       string
	}
	var qList = []testCase{
		{
			"INSERT INTO doc_type (doc) VALUES ('passport');",
			"passport",
			"query INSERT INTO doc_type (doc) VALUES ('passport'); is not a SELECT query",
		},
		{
			"SELECT doc from doc_type WHERE doc=?",
			"Test Passport",
			"",
		},
		{
			"SELECT doc from doc_type WHERE doc=?",
			"passport",
			"",
		},
		{
			"SELECT doc from doc_type WHERE doc=?",
			"veterinary passport",
			"",
		},
		{
			"SELECT doc from doc_type WHERE doc=?",
			"military passport",
			"",
		},
		{
			"SELECT type from animal_type WHERE type=?",
			"dog",
			"",
		},
		{
			"SELECT type from animal_type WHERE type=?",
			"cat",
			"",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbTimeout)*time.Millisecond)
	defer cancel()
	var doc string

	for _, val := range qList {
		rows, err := db.Get(ctx, val.query, val.valideVal)
		if err != nil {
			if err.Error() == val.err {
				continue
			} else {
				t.Fatalf("unexpected error: %s", err.Error())
			}
		}

		rows.Next()
		err = rows.Scan(&doc)
		if err != nil {
			t.Fatalf("Failed to read cursor: %s", err.Error())
		}

		if doc != val.valideVal {
			t.Fatalf("got value [%s] insted of [passport]", doc)
		}
	}
}

func TestSqLiteDB_Close(t *testing.T) {
	t.Cleanup(deleteFile)
	if err := db.Close(); err != nil {
		t.Fatalf("Can not clode DB connection: %s", err)
	}
}

func deleteFile() {
	os.Remove(testDbUri)
}
