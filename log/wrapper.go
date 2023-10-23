package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// Wrapper implements various useful log functions with the provided logger
// to fulfill interfaces.
type Wrapper struct {
	logger *slog.Logger
	level  slog.Level
}

func NewWrapper(logger *slog.Logger, level slog.Level) *Wrapper {
	return &Wrapper{
		logger: logger,
		level:  level,
	}
}

func (w *Wrapper) Printf(template string, args ...interface{}) {
	if !w.logger.Enabled(context.Background(), w.level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), w.level, fmt.Sprintf(template, args...), pcs[0])
	_ = w.logger.Handler().Handle(context.Background(), r)
}
