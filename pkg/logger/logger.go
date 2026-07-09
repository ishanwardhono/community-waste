package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var l = slog.New(slog.NewTextHandler(os.Stdout, nil))

func Init(level string) {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
}

func Infof(ctx context.Context, format string, args ...any) {
	l.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func Errorf(ctx context.Context, format string, args ...any) {
	l.ErrorContext(ctx, fmt.Sprintf(format, args...))
}

func Fatalf(ctx context.Context, format string, args ...any) {
	l.ErrorContext(ctx, fmt.Sprintf(format, args...))
	os.Exit(1)
}
