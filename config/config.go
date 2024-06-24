package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"slices"
)

type Config struct {
	DB struct {
		Type    string `yaml:"type" env-default:"sqlite" env-description:"Database type. Allowed pgsql, sqlite" validate:"required,oneof=pgsql sqlite"`
		Uri     string `yaml:"uri" env-required:"true" env-description:"Database URI. Example postgresql://user:secret@host:5432/repos" validate:"required,uri"`
		Timeout int    `yaml:"timeout" env-default:"1000" env-description:"Timeout for an SQL query" validate:"required,number,gt=0"`
		InitDB  bool   `yaml:"initDB" env-default:"false" env-description:"Init DB with default schema"`
	} `yaml:"db"`
	Web struct {
		Port        int `yaml:"port" env-default:"8080" env-description:"default server port" validate:"required,number,gt=79"`
		Timeout     int `yaml:"timeout" env-default:"4000" env-description:"Connection timeout" validate:"required,number,gt=0"`
		IdleTimeout int `yaml:"idleTimeout" env-default:"60000" env-description:"Idle connection timeout" validate:"required,number,gt=0"`
	} `yaml:"web"`
	Log struct {
		Level  string `yaml:"level" env-default:"error" env-description:"App logLevel. Allowed debug, info, warn, error" validate:"required,oneof=debug info warn error"`
		Format string `yaml:"format" env-default:"text" env-description:"App log format. Allowed text, json" validate:"required,oneof=text json"`
	} `yaml:"log"`
}

func GetConfPath() (string, error) {
	var path string
	validate := validator.New(validator.WithRequiredStructEnabled())
	i := slices.IndexFunc(os.Args, func(s string) bool {
		return s == "--config"
	})
	// i+2 means something ahead of --config arg exists
	if i == -1 || len(os.Args) < i+2 {
		return "", fmt.Errorf("--config parameter not presented or empty")
	}
	path = os.Args[i+1]
	if err := validate.Var(path, "required,file"); err != nil {
		return "", fmt.Errorf("config path %s either not exists or not legal: %w", path, err)
	}
	return path, nil
}

func (c *Config) New(path string) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := cleanenv.ReadConfig(path, c); err != nil {
		return err
	}
	if err := validate.Struct(c); err != nil {
		return err
	}
	return nil
}
