package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var l *slog.Logger

func init() {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "production" || env == "prod" {
		l = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
}

func Info(ctx context.Context, msg string, args ...any)  { l.InfoContext(ctx, msg, args...) }
func Warn(ctx context.Context, msg string, args ...any)  { l.WarnContext(ctx, msg, args...) }
func Error(ctx context.Context, msg string, args ...any) { l.ErrorContext(ctx, msg, args...) }
