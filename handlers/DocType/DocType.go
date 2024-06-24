package DocType

// This code developed for learning purposes
// Filling look-up tables will be done with init action
// And this handler will be omitted

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"mis-catanddog/repos"
	"net/http"
)

// DocType handles CRUD operation for the /doc_type url.
// It receives DB object of type interfaces.DB from the request context.
func DocType(w http.ResponseWriter, r *http.Request) {
	// get logger
	log, ok := (r.Context().Value("logger")).(*slog.Logger)
	if !ok {
		// how to log without a logger?
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log = log.With("ID", uuid.New())

	log.Info("request", "Method", r.Method, "Host", r.Host, "URL", r.URL, "Headers", r.Header)

	// get repo
	db, ok := (r.Context().Value("db")).(repos.DB)
	if !ok {
		log.Error("cannot get DB object from context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// select handler
	switch r.Method {
	case http.MethodGet:
		getDocType(r.Context(), w, r, log, db)
	default:
		log.Error(fmt.Sprintf("unexpected method %s", r.Method))
	}
}
