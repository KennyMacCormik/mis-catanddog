package handlers

// This code developed for learning purposes
// Filling look-up tables will be done with init action
// And this handler will be omitted

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"mis-catanddog/config"
	"mis-catanddog/interfaces"
	"mis-catanddog/lg"
	"net/http"
	"sync"
	"time"
)

// DocType handles CRUD operation for the /doc_type url.
// It receives DB object of type interfaces.DB from the request context.
// It panics in case it can't.
func DocType(w http.ResponseWriter, r *http.Request) {
	log := lg.Logger.With("ID", uuid.New())
	log.Info("request", "Method", r.Method, "Host", r.Host, "URL", r.URL, "Headers", r.Header)
	db, ok := (r.Context().Value("DB")).(interfaces.DB)
	if !ok {
		log.Error("PANIC cannot get DB object")
		panic("cannot get DB object")
	}
	switch r.Method {
	case http.MethodGet:
		getDocType(r.Context(), w, r, log, db)
	/*
		case http.MethodPost:
				postDocType(r.Context(), w, r, log, db)
		case http.MethodDelete:
			deleteDocType(r.Context(), w, r, log, db)
		case http.MethodPut, http.MethodPatch:
			updateDocType(r.Context(), w, r, log, db)
	*/
	default:
		log.Error(fmt.Sprintf("unexpected method %s", r.Method))
	}
}

func postDocType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate request headers and other stuff
	if err := validateContentType(w, r, l); err != nil {
		return
	}

	// decode body to json
	var body []any // I need to pass each value to variadic
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		l.Error(fmt.Errorf("cannot decode json: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	l.Debug("json.NewDecoder result", "body", fmt.Sprintf("%v", body))

	// prep query
	lenBody := len(body)
	placeHolders := ""
	for i := 0; i < lenBody; i++ {
		placeHolders += "(?), "
	}
	placeHolders = placeHolders[:len(placeHolders)-2]

	// run query
	dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
	defer cancel()
	q := "INSERT INTO doc_type (doc) VALUES "
	l.Debug("query", "base", "["+q+"]", "placeHolders", "["+placeHolders+"]", "body", body)
	err := db.Exec(dbCtx, q+placeHolders, body...)
	if err != nil {
		l.Error(fmt.Errorf("bad request: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func deleteDocType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate query
	q := r.URL.Query()
	val, ok := q["id"]
	if !ok {
		l.Error("id not present in query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	query := "DELETE FROM doc_type WHERE id IN"

	// prep query
	lenBody := len(val)
	placeHolders := "("
	for i := 0; i < lenBody; i++ {
		placeHolders += "?, "
	}
	placeHolders = placeHolders[:len(placeHolders)-2]
	placeHolders += ")"
	query += placeHolders

	// prep args for variadic
	var args []any
	for _, vl := range val {
		args = append(args, vl)
	}

	// run query
	dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
	defer cancel()
	l.Debug("query", "base", query)
	err := db.Exec(dbCtx, query, args...)
	if err != nil {
		l.Error(fmt.Errorf("bad request: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// TODO fix the thing
func updateDocType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate query
	q := r.URL.Query()
	val, ok := q["id"]
	if !ok {
		l.Error("id not present in query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// decode body to json
	body := make(map[string]string) // I need to pass each value to variadic
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		l.Error(fmt.Errorf("cannot decode json: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	l.Debug("prepare for query", "body", body, "val", val)

	// run query
	var wg sync.WaitGroup
	var failed []string
	var m sync.Mutex

	for _, vl := range val {
		tmp, ok := body[vl]
		if !ok {
			l.Error(fmt.Sprintf("bad request. value for [%s] is missing in body", vl))
			failed = append(failed, vl)
			continue
		}
		query := "UPDATE doc_type SET doc=? WHERE id=?"
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
			defer cancel()
			err := db.Exec(dbCtx, query, tmp, vl) // go 1.22 fine with this?
			if err != nil {
				m.Lock()
				failed = append(failed, vl) // Is this safe? it requires mutex?
				m.Unlock()
				l.Error(fmt.Errorf("bad request: %w", err).Error())
				return
			}
		}()
	}
	wg.Wait()
	if len(failed) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		// return to caller
		jsonString, err := json.Marshal(failed)
		if err != nil {
			l.Error("failed to marshal json", "error", fmt.Errorf("failed to marshal json: %w", err).Error())
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, string(jsonString)); err != nil {
			l.Error(fmt.Errorf("cannot write responce to caller: %w", err).Error())
			return
		}
	}
}
