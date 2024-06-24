package controllers

import (
	"context"
	"log/slog"
)

type DocType struct {
	Id  int
	Doc string
	Err string
}

type DocTypeGetter interface {
	DocTypeGetById(ctx context.Context, id int, l *slog.Logger) DocType
	DocTypeGetByDoc(ctx context.Context, doc string, l *slog.Logger) DocType
}
