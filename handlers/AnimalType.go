package handlers

// This code developed for learning purposes
// Filling look-up tables will be done with init action
// And this handler will be omitted

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"mis-catanddog/config"
	"mis-catanddog/interfaces"
	"mis-catanddog/lg"
	"net/http"
	"strconv"
	"time"
)

// AnimalType handles CRUD operation for the /animal_type url.
// It receives DB object of type interfaces.DB from the request context.
// It panics in case it can't.
func AnimalType(w http.ResponseWriter, r *http.Request) {
	log := lg.Logger.With("ID", uuid.New())
	log.Info("request", "Method", r.Method, "Host", r.Host, "URL", r.URL, "Headers", r.Header)
	db, ok := (r.Context().Value("DB")).(interfaces.DB)
	if !ok {
		log.Error("PANIC cannot get DB object")
		panic("cannot get DB object")
	}
	switch r.Method {
	case http.MethodGet:
		getAnimalType(r.Context(), w, r, log, db)
	case http.MethodPost:
		postAnimalType(r.Context(), w, r, log, db)
	//case http.MethodDelete:
	//deleteDocType(r.Context(), w, r, log, db)
	//case http.MethodPut, http.MethodPatch:
	//updateDocType(r.Context(), w, r, log, db)
	default:
		log.Error(fmt.Sprintf("unexpected method %s", r.Method))
	}
}

func getAnimalType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate query
	q := r.URL.Query()
	valType, okType := q["type"]
	valId, okId := q["id"]
	if (!okType && !okId) || (okType && okId) {
		l.Error("ambiguous query. 'type' and 'id' either together or not present", "query", q)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var val []string
	if okType {
		val = valType
	} else {
		val = valId
	}
	l.Debug("query content", "val", val)

	if len(val) > 1 {
		l.Error(fmt.Sprintf("unexpected number of '%s' queries", val[0]), "query", val)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res := val[0]

	// execute query
	dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
	defer cancel()
	var qResult *sql.Rows
	var err error
	if okType {
		qResult, err = db.Get(dbCtx, "SELECT id, type from animal_type WHERE type=?", res)
	} else {
		qResult, err = db.Get(dbCtx, "SELECT id, type from animal_type WHERE id=?", res)
	}
	if err != nil {
		l.Error(fmt.Errorf("bad request: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer qResult.Close()
	if qResult.Err() != nil {
		l.Debug("query result", "error", qResult.Err().Error())
	}

	// parse result
	id, animalType := make([]int, 0), make([]string, 0)
	for qResult.Next() {
		var s string
		var i int
		if err := qResult.Scan(&i, &s); err != nil {
			l.Error(fmt.Errorf("cannot read query result %w", err).Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id = append(id, i)
		animalType = append(animalType, s)
	}
	if qResult.Err() != nil {
		l.Debug("query result", "error", qResult.Err().Error())
	}
	l.Debug("query result", "id", id, "animal_type", animalType)

	// prep return value
	tmpLen := len(id)
	returnJson := "{"
	for i := 0; i < tmpLen; i++ {
		returnJson += strconv.Itoa(id[i]) + ":" + animalType[i] + ","
	}
	if returnJson != "{" {
		returnJson = returnJson[:len(returnJson)-1]
	}
	returnJson += "}"

	// return to caller
	w.Header().Set("Content-Type", "application/json")
	if _, err = io.WriteString(w, returnJson); err != nil {
		l.Error(fmt.Errorf("cannot write responce to caller: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func postAnimalType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate request headers and other stuff
	if err := validateContentType(w, r, l); err != nil {
		return
	}

	// decode body to json
	var body []any // I need to pass each value to variadic. Try type assertion to ensure it is string?
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
	q := "INSERT INTO animal_type (type) VALUES "
	l.Debug("query", "base", "["+q+"]", "placeHolders", "["+placeHolders+"]", "body", body)
	err := db.Exec(dbCtx, q+placeHolders, body...)
	if err != nil {
		l.Error(fmt.Errorf("bad request: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
