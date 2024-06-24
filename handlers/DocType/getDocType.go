package DocType

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mis-catanddog/controllers"
	"mis-catanddog/repos"
	"net/http"
	"net/url"
	"strconv"
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

// getDocTypeQueryData returns results to hand over to client
func getDocTypeQueryData(ctx context.Context, vals url.Values, db controllers.DocTypeGetter, l *slog.Logger) []controllers.DocType {
	var result []controllers.DocType
	valDoc, okDoc := vals["doc"]
	if okDoc {
		for _, val := range valDoc {
			result = append(result, db.DocTypeGetByDoc(ctx, val, l))
		}
	} else {
		valId, _ := vals["id"]
		for _, val := range valId {
			intVal, err := strconv.Atoi(val)
			if err != nil {
				result = append(result, controllers.DocType{Id: 0, Doc: "", Err: fmt.Sprintf("failed to convert id [%s] to an integer", val)})
				return result
			}
			result = append(result, db.DocTypeGetById(ctx, intVal, l))
		}
	}
	return result
}

// getDocTypeHideInternals hides any possible error details
func getDocTypeHideInternals(result []controllers.DocType, l *slog.Logger) {
	ln := len(result)
	for i := 0; i < ln; i++ {
		if result[i].Err != "" {
			result[i].Err = "Bad request"
		}
		// id = 0 means empty result for the query
		if result[i].Id == 0 {
			result[i].Err = "Empty result"
		}
	}
}

func getDocType(ctx context.Context, w http.ResponseWriter, r *http.Request, l *slog.Logger, db repos.DB) {
	// validate URL query
	if err := getDocTypeValidateUrl(r.URL.Query()); err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// convert to controller
	controller, ok := db.(controllers.DocTypeGetter)
	if !ok {
		l.Error("object of type [DB] interface failed to covert to [DocTypeGetter] interface")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// get results
	result := getDocTypeQueryData(ctx, r.URL.Query(), controller, l)

	// return to caller
	getDocTypeHideInternals(result, l)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		l.Error(fmt.Errorf("cannot write responce to caller: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
