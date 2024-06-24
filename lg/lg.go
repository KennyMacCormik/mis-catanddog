package lg

import (
	"fmt"
	"log/slog"
	"os"
)

func Init(format, level string) (*slog.Logger, error) {
	var levelMap = map[string]slog.Level{
		"debug": -4,
		"info":  0,
		"warn":  4,
		"error": 8,
	}

	switch format {
	case "json":
		h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: levelMap[level]})
		l := slog.New(h)
		return l, nil
	case "text":
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: levelMap[level]})
		l := slog.New(h)
		return l, nil
	default:
		return nil, fmt.Errorf("unexpected log format %s. Expected: text, json", format)
	}
}
