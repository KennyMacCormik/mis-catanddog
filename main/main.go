package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"mis-catanddog/config"
	"mis-catanddog/handlers/DocType"
	"mis-catanddog/lg"
	"mis-catanddog/repos"
	"mis-catanddog/repos/sqlite3"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	var db repos.DB
	var cfg config.Config
	var logg *slog.Logger
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// read config
	path, err := config.GetConfPath()
	if err != nil {
		log.Fatal(fmt.Errorf("config path error: %w", err))
	}
	if err := cfg.New(path); err != nil {
		log.Fatal(fmt.Errorf("reading config file: %w", err))
	}

	// init logger
	logg, err = lg.Init(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		log.Fatal(fmt.Errorf("logger init error: %w", err))
	}

	logg.Debug(fmt.Sprintf("Config: %+v", cfg))

	// create DB connection
	db = initRepo(cfg, logg)
	if db == nil {
		os.Exit(1)
	}
	defer db.Close()

	// init server
	server := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Web.Port),
		Handler:        nil,
		ReadTimeout:    time.Duration(cfg.Web.Timeout) * time.Millisecond,
		WriteTimeout:   time.Duration(cfg.Web.Timeout) * time.Millisecond,
		IdleTimeout:    time.Duration(cfg.Web.IdleTimeout) * time.Millisecond,
		MaxHeaderBytes: 1 << 20, // 1Mb
		BaseContext: func(l net.Listener) context.Context {
			// Cant find parent context in documentation
			ctx := context.WithValue(context.TODO(), "db", db)
			return context.WithValue(ctx, "logger", logg)
		},
	}
	http.HandleFunc("/doc_type", DocType.DocType)

	logg.Info("Starting server")
	err = server.ListenAndServe()
	if err != nil {
		logg.Error(fmt.Errorf("web server failed: %w", err).Error())
		os.Exit(1)
	}
}

func initRepo(cfg config.Config, l *slog.Logger) repos.DB {
	switch cfg.DB.Type {
	case "sqlite":
		var db repos.DB = &sqlite3.SqLiteDB{}
		u, err := url.ParseRequestURI(cfg.DB.Uri)
		if err != nil {
			l.Error(fmt.Errorf("failed to parce uri: %w", err).Error())
			return nil
		}
		_, err = os.Stat(u.Opaque)
		if err != nil {
			l.Error(fmt.Errorf("sqlite db file does not exist: %w", err).Error())
			return nil
		}
		err = db.New(cfg.DB.Uri, time.Duration(cfg.DB.Timeout)*time.Millisecond)
		if err != nil {
			l.Error(fmt.Errorf("repos connection error: %w", err).Error())
			return nil
		}
		return db
	case "pgsql":
		l.Error("not yet implemented")
		return nil
	default:
		l.Error("unexpected db type")
		return nil
	}
}
