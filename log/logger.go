package log

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	Color    bool   `conf:"color"`
}

// InitLogger loads a global logger based on a configuration
func InitLogger(c *LoggerConfig) error {
	level, err := ParseLogLevel(c.Level)
	if err != nil {
		return err
	}

	var logger *slog.Logger
	switch c.Encoding {
	case EncodingText, "console":
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					if source, ok := a.Value.Any().(*slog.Source); ok {
						a.Value = slog.StringValue(TrimSource(source.File, 2) + ":" + strconv.Itoa(source.Line))
					}
				}
				if a.Key == slog.LevelKey && c.Color {
					if level, ok := a.Value.Any().(slog.Level); ok {
						switch level {
						case slog.LevelDebug:
							a.Value = slog.StringValue("DEBUGGGG")
						}
					}
				}
				return a
			},
		}))
	case EncodingJSON:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					if source, ok := a.Value.Any().(*slog.Source); ok {
						a.Value = slog.StringValue(TrimSource(source.File, 2) + ":" + strconv.Itoa(source.Line))
					}
				}
				return a
			},
		}))
	}
	Logger = logger
	slog.SetDefault(logger)
	return nil
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
