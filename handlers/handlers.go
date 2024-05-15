package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
)

// validateUrl validetes query URL according to the required logic. Valid modes are
// oneOf: returns values that exists in query from all the available in an array. In case more than one available returns error
func validateUrl(q url.Values, mode string, array ...string) {

}

// validateRequest checks all necessary mumbo-jumbo. In case any errors it logs them, sets http.StatusBadRequest
// and returns "" error as a sign that request is bad. Returns nil in case all is fine.
func validateContentType(w http.ResponseWriter, r *http.Request, l *slog.Logger) error {
	val, ok := r.Header["Content-Type"]
	if !ok {
		l.Error("content type not set")
		w.WriteHeader(http.StatusBadRequest)
		return fmt.Errorf("")
	}
	if slices.IndexFunc(val, func(s string) bool { return s == "application/json" }) < 0 {
		l.Error(fmt.Sprintf("unsupprted content types %v", val))
		w.WriteHeader(http.StatusBadRequest)
		return fmt.Errorf("")
	}
	return nil
}
