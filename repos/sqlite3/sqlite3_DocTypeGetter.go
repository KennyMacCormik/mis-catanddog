package sqlite3

import (
	"context"
	"fmt"
	"log/slog"
	"mis-catanddog/controllers"
	"mis-catanddog/repos"
)

// DocTypeGetById searches DocType table by id and returns DocType object
func (s *SqLiteDB) DocTypeGetById(ctx context.Context, id int, l *slog.Logger) controllers.DocType {
	req := repos.DbReq{Query: "SELECT id, doc from doc_type WHERE id=?", Args: append(make([]any, 0), id)}

	return invokeRequest(ctx, req, s, l)
}

// DocTypeGetByDoc searches DocType table by doc and returns DocType object
func (s *SqLiteDB) DocTypeGetByDoc(ctx context.Context, doc string, l *slog.Logger) controllers.DocType {
	req := repos.DbReq{Query: "SELECT id, doc from doc_type WHERE doc=?", Args: append(make([]any, 0), doc)}

	return invokeRequest(ctx, req, s, l)
}

func invokeRequest(ctx context.Context, req repos.DbReq, s *SqLiteDB, l *slog.Logger) controllers.DocType {
	var result controllers.DocType

	rows, err := s.Get(ctx, req)
	if err != nil {
		tmp := controllers.DocType{Id: 0, Doc: "", Err: fmt.Errorf("bad DB query: %w", err).Error()}
		l.Error(tmp.Err)
		return result
	}

	for i := 0; rows.Next(); i++ {
		if i > 0 {
			l.Error("query to dict table yielded more than one result")
			result = controllers.DocType{Id: 0, Doc: "", Err: "query to dict table yielded more than one result"}
			break
		}
		if err := rows.Scan(&result.Id, &result.Doc); err != nil {
			result.Err = fmt.Errorf("cannot read query result %w", err).Error()
			continue
		}
	}
	l.Debug("query result", "id", result.Id, "doc_type", result.Doc, "error", result.Err)

	return result
}
