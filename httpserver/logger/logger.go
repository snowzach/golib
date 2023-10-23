package logger

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Level        slog.Level `conf:"level"`
	RequestBody  bool       `conf:"request_body"`
	ResponseBody bool       `conf:"response_body"`
	IgnorePaths  []string   `conf:"ignore_paths"`
}

func LoggerStandardMiddleware(logger *slog.Logger, config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the prefix should be ignored
			for _, prefix := range config.IgnorePaths {
				if strings.HasPrefix(r.RequestURI, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()

			// See if we're already using a wrapped response writer and if not make one.
			ww, ok := w.(middleware.WrapResponseWriter)
			if !ok {
				ww = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			}

			// If we should log bodies, setup buffer to capture
			var responseBody *bytes.Buffer
			if config.ResponseBody {
				responseBody = new(bytes.Buffer)
				ww.Tee(responseBody)
			}

			next.ServeHTTP(ww, r)

			// If the remote IP is being proxied, use the real IP
			remoteIP := r.Header.Get("x-real-ip")
			if remoteIP == "" {
				remoteIP = r.Header.Get("x-forwarded-for")
				if remoteIP == "" {
					remoteIP = r.RemoteAddr
				}
			}

			fields := []slog.Attr{
				{Key: "status", Value: slog.IntValue(ww.Status())},
				{Key: "duration", Value: slog.DurationValue(time.Since(start))},
				{Key: "path", Value: slog.StringValue(r.RequestURI)},
				{Key: "method", Value: slog.StringValue(r.Method)},
				{Key: "protocol", Value: slog.StringValue(r.Proto)},
				{Key: "agent", Value: slog.StringValue(r.UserAgent())},
				{Key: "remote", Value: slog.StringValue(remoteIP)},
			}

			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				fields = append(fields, slog.Attr{Key: "request-id", Value: slog.StringValue(reqID)})
			}

			if config.RequestBody {
				if req, err := httputil.DumpRequest(r, true); err == nil {
					fields = append(fields, slog.Attr{Key: "request", Value: slog.StringValue(string(req))})
				}
			}
			if config.ResponseBody {
				fields = append(fields, slog.Attr{Key: "response", Value: slog.StringValue(responseBody.String())})
			}

			// Write the log entry assuming we're logging at that level.
			logger.LogAttrs(r.Context(), config.Level, "HTTP Request", fields...)
		})
	}
}
