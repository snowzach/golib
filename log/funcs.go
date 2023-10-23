package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func Debug(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelDebug, 3, msg, args...)
}

func Debugf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelDebug, 3, fmt.Sprintf(template, args...))
}

func Info(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelInfo, 3, msg, args...)
}

func Infof(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelInfo, 3, fmt.Sprintf(template, args...))
}

func Warn(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelWarn, 3, msg, args...)
}

func Warnf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelWarn, 3, fmt.Sprintf(template, args...))
}

func Error(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, msg, args...)
}

func Errorf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, fmt.Sprintf(template, args...))
}

func Fatal(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, msg, args...)
	os.Exit(1)
}

func Fatalf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, fmt.Sprintf(template, args...))
	os.Exit(1)
}

func Panic(msg string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, msg, args...)
	panic(nil)
}

func Panicf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelError, 3, fmt.Sprintf(template, args...))
	panic(nil)
}

func Println(msg string) {
	LogSkip(context.Background(), Logger, slog.LevelInfo, 3, msg)
}

func Printf(template string, args ...interface{}) {
	LogSkip(context.Background(), Logger, slog.LevelInfo, 3, fmt.Sprintf(template, args...))
}

func LogSkip(ctx context.Context, logger *slog.Logger, level slog.Level, skip int, msg string, args ...interface{}) {
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(context.Background(), r)
}
