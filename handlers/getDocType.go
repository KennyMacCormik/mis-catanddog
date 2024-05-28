package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"mis-catanddog/config"
	"mis-catanddog/interfaces"
	"net/http"
	"net/url"
	"time"
)

// getDocTypeValidateUrl validates request URL
func getDocTypeValidateUrl(vals url.Values) error {
	_, okDoc := vals["doc"]
	_, okId := vals["id"]
	if (!okDoc && !okId) || (okDoc && okId) {
		return fmt.Errorf("ambiguous query; 'doc' and 'id' either together or not present; query [%s]", vals)
	}
	return nil
}

// getDocTypePrepData returns IDs to query and query to run
func getDocTypePrepData(vals url.Values) ([]string, string) {
	valDoc, okDoc := vals["doc"]
	valId, _ := vals["id"]
	if okDoc {
		return valDoc, "SELECT id, doc from doc_type WHERE doc=?"
	} else {
		return valId, "SELECT id, doc from doc_type WHERE id=?"
	}
}

// getDocTypeParseResult reads result and returns it.
// Functions panics if there are more than one query result
func getDocTypeParseResult(rows *sql.Rows, l *slog.Logger) docType {
	var result docType
	for i := 0; rows.Next(); i++ {
		if i > 0 {
			l.Error("PANIC. query to dict table yielded more than one result")
			panic("PANIC. query to dict table yielded more than one result")
		}
		if err := rows.Scan(&result.Id, &result.Doc); err != nil {
			result.Err = fmt.Errorf("cannot read query result %w", err).Error()
			continue
		}
	}
	l.Debug("query result", "id", result.Id, "doc_type", result.Doc)
	return result
}

// getDocTypeExecQuery executes separate query to DB for each ID provided
func getDocTypeExecQuery(ctx context.Context, db interfaces.DB, l *slog.Logger, ids []string, query string) []docType {
	var result []docType
	for _, val := range ids {
		dbCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Cfg.DB.Timeout)*time.Millisecond)
		defer cancel()
		qResult, err := db.Get(dbCtx, query, val)
		if err != nil {
			tmp := docType{0, "", fmt.Errorf("bad DB query: %w", err).Error()}
			result = append(result, tmp)
			l.Error(tmp.Err)
			continue
		}
		result = append(result, getDocTypeParseResult(qResult, l))
		qResult.Close()
	}
	return result
}

// getDocTypeHideInternals hides any possible error details
func getDocTypeHideInternals(result []docType) {
	ln := len(result)
	for i := 0; i < ln; i++ {
		// id = 0 mean empty query. It only happens on bad request
		if result[i].Err != "" || result[i].Id == 0 {
			result[i].Err = "Bad request"
		}
	}
}

func getDocType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db interfaces.DB) {
	// validate URL query
	if err := getDocTypeValidateUrl(r.URL.Query()); err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// prep data for query
	queryData, query := getDocTypePrepData(r.URL.Query())

	// execute query
	result := getDocTypeExecQuery(ctx, db, l, queryData, query)

	// return to caller
	getDocTypeHideInternals(result)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		l.Error(fmt.Errorf("cannot write responce to caller: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
