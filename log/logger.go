package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

// Defaults
var (
	// Sane default logger setup
	Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}))

	ErrUnknownLogLevel    = errors.New("unknown log level")
	ErrUnknownLogEncoding = errors.New("unknown log encoding")
)

const (
	EncodingText = "text"
	EncodingJSON = "json"
)

type LoggerConfig struct {
	Level    string `conf:"level"`
	Encoding string `conf:"encoding"`
	Color    bool   `conf:"color"` // Only valid for console encoding.
	Output   string `conf:"output"`
}

// InitLogger loads a global logger based on a configuration
func InitLogger(c *LoggerConfig) error {
	// Parse the level
	level, err := ParseLogLevel(c.Level)
	if err != nil {
		return err
	}

	// Determine the output
	var w io.Writer
	switch c.Output {
	case "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default: // Otherwise assume it's a log file path
		if w, err = os.Open(c.Output); err != nil {
			return fmt.Errorf("log output file open error: %w", err)
		}
	}

	var logger *slog.Logger
	switch c.Encoding {
	case EncodingText, "console":
		if c.Color {
			logger = slog.New(tint.NewHandler(w, &tint.Options{
				AddSource:   true,
				Level:       level,
				ReplaceAttr: ReplaceAttrTrimSource,
				TimeFormat:  time.RFC3339Nano,
			}))
		} else {
			logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
				AddSource:   true,
				Level:       level,
				ReplaceAttr: ReplaceAttrTrimSource,
			}))
		}
	case EncodingJSON:
		logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       level,
			ReplaceAttr: ReplaceAttrTrimSource,
		}))
	}
	Logger = logger
	slog.SetDefault(logger)
	return nil
}

func ReplaceAttrTrimSource(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if source, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(TrimSource(source.File, 2) + ":" + strconv.Itoa(source.Line))
		}
	}
	return a
}

func TrimSource(source string, parts int) string {
	i := len(source) - 1
	for ; i > 0; i-- {
		if source[i] == os.PathSeparator {
			parts--
			if parts == 0 {
				return source[i+1:]
			}
		}
	}
	return source
}

// ParseLogLevel is used to parse configuration options into a log level.
func ParseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "err", "error":
		return slog.LevelError, nil
	default:
		return 0, ErrUnknownLogLevel
	}
}
