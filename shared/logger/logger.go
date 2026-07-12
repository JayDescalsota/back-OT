package logger

import (
	"context"
	"log/slog"
	"os"
)

var l = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func Info(ctx context.Context, msg string, args ...any)  { l.InfoContext(ctx, msg, args...) }
func Warn(ctx context.Context, msg string, args ...any)  { l.WarnContext(ctx, msg, args...) }
func Error(ctx context.Context, msg string, args ...any) { l.ErrorContext(ctx, msg, args...) }
