package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"mis-catanddog/config"
	"mis-catanddog/interfaces"
	"mis-catanddog/lg"
	"net/http"
	"net/url"
	"time"
)

// HumanSearch handles search operation for the /human/{$} url.
// It receives DB object of type interfaces.DB from the request context.
// It panics in case it can't.
func HumanSearch(w http.ResponseWriter, r *http.Request) {
	log := lg.Logger.With("ID", uuid.New())
	log.Info("request", "Method", r.Method, "Host", r.Host, "URL", r.URL, "Headers", r.Header)
	db, ok := (r.Context().Value("DB")).(interfaces.DB)
	if !ok {
		log.Error("PANIC cannot get DB object")
		panic("cannot get DB object")
	}
	getHumanSearch(r.Context(), w, r, log, db)
}

// HumanId handles CRUD operation for the /human/id/{id}/{$} url.
// It receives DB object of type interfaces.DB from the request context.
// It panics in case it can't.
func HumanId(w http.ResponseWriter, r *http.Request) {
	log := lg.Logger.With("ID", uuid.New())
	log.Info("request", "Method", r.Method, "Host", r.Host, "URL", r.URL, "Headers", r.Header)
	db, ok := (r.Context().Value("DB")).(interfaces.DB)
	if !ok {
		log.Error("PANIC cannot get DB object")
		panic("cannot get DB object")
	}
	switch r.Method {
	case http.MethodGet:
		getHumanId(r.Context(), w, r, log, db)
	//case http.MethodPost:
	//postDocType(r.Context(), w, r, log, db)
	//case http.MethodDelete:
	//deleteDocType(r.Context(), w, r, log, db)
	//case http.MethodPut, http.MethodPatch:
	//updateDocType(r.Context(), w, r, log, db)
	default:
		log.Error(fmt.Sprintf("unexpected method %s", r.Method))
	}
}

func getHuman(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {

}

func getHumanId(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
}

func getHumanSearch(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	q, a, err := prepSearchQuery(r.URL)
	if err != nil {
		l.Error(fmt.Errorf("cannot prepare sql query: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
	defer cancel()
	var qResult *sql.Rows
	qResult, err = db.Get(dbCtx, q, a...)
	if err != nil {
		l.Error(fmt.Errorf("bad request: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer qResult.Close()
	if qResult.Err() != nil {
		l.Debug("query result", "error", qResult.Err().Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// parse result
	id, docType := make([]int, 0), make([]string, 0)
	for qResult.Next() {
		var doc_id, firsst_name, middle_name, last_name string
		var doc_type int
		var birth_date time.Time
		if err := qResult.Scan(&doc_id, &doc_type, &firsst_name, &middle_name, &last_name, &birth_date); err != nil {
			l.Error(fmt.Errorf("cannot read query result %w", err).Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if qResult.Err() != nil {
		l.Debug("query result", "error", qResult.Err().Error())
	}
	l.Debug("query result", "id", id, "doc_type", docType)
}

func prepSearchQuery(u *url.URL) (query string, args []any, err error) {
	var a []any
	var result, q string
	v := u.Query()
	// handle search by doc_id
	if val, ok := v["doc_id"]; ok {
		cols := "doc_id, doc_type, firsst_name, middle_name, last_name, birth_date"
		for _, val1 := range val {
			a = append(a, val1)
			q += "doc_id=?, "
		}
		result = "SELECT " + cols + " FROM human WHERE "
		result += q[:len(q)-2]
		return result, a, nil
	}
	// handle complex search that yields multiple result
	return "", nil, nil
}
