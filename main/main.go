package main

import (
	"context"
	"fmt"
	"log"
	"mis-catanddog/config"
	"mis-catanddog/database"
	"mis-catanddog/handlers"
	"mis-catanddog/interfaces"
	"mis-catanddog/lg"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// read config
	path, err := config.GetConfPath()
	if err != nil {
		log.Fatal(fmt.Errorf("config path error: %w", err))
	}
	if err := config.Cfg.New(path); err != nil {
		log.Fatal(fmt.Errorf("reading config file: %w", err))
	}

	// init logger
	lg.Logger, err = lg.Init(config.Cfg.Log.Format, config.Cfg.Log.Level)
	if err != nil {
		log.Fatal(fmt.Errorf("logger init error: %w", err))
	}

	lg.Logger.Debug(fmt.Sprintf("Config: %+v", config.Cfg))

	// create DB connection
	var db interfaces.DB
	switch config.Cfg.DB.Type {
	case "sqlite":
		db = &database.SqLiteDB{}
		err = db.New(config.Cfg.DB.Uri, config.Cfg.DB.Type)
		if err != nil {
			lg.Logger.Error(fmt.Errorf("database connection error: %w", err).Error())
			os.Exit(1)
		}
		defer db.Close()
		lg.Logger.Info("DB connection successful")
	case "pgsql":
		lg.Logger.Error("not yet implemented")
		os.Exit(1)
	default:
		lg.Logger.Error("unexpected db type")
		os.Exit(1)
	}

	// init DB if empty
	initDB(db)

	// init server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    time.Duration(config.Cfg.Web.Timeout) * time.Millisecond,
		WriteTimeout:   time.Duration(config.Cfg.Web.Timeout) * time.Millisecond,
		IdleTimeout:    time.Duration(config.Cfg.Web.IdleTimeout) * time.Millisecond,
		MaxHeaderBytes: 1 << 20, // 1Mb
		BaseContext: func(l net.Listener) context.Context {
			return context.WithValue(context.TODO(), "DB", db)
		}, // Cant find parent context in documentation
	}
	//http.HandleFunc("/doc_type", handlers.DocType)
	//http.HandleFunc("/animal_type", handlers.AnimalType)
	http.HandleFunc("GET /human/{$}", handlers.HumanSearch)
	http.HandleFunc("/human/id/{id}/{$}", handlers.HumanId)
	lg.Logger.Info("Starting server")
	err = server.ListenAndServe()
	if err != nil {
		lg.Logger.Error(fmt.Errorf("web server failed: %w", err).Error())
		os.Exit(1)
	}
}

func initDB(db interfaces.DB) {
	if config.Cfg.DB.InitDB {
		if err := db.Init(time.Duration(config.Cfg.DB.Timeout)); err != nil {
			lg.Logger.Error(fmt.Errorf("failed to init DB: %w", err).Error())
			os.Exit(1)
		}
		lg.Logger.Info("DB init successful")
	} else {
		lg.Logger.Debug("DB init disabled")
	}
}
