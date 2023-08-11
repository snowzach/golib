package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap/zapcore"

	"github.com/snowzach/golib/httpserver"
	"github.com/snowzach/golib/httpserver/logger"
	"github.com/snowzach/golib/httpserver/metrics"
	"github.com/snowzach/golib/log"
	"github.com/snowzach/golib/signal"
)

func main() {

	if err := log.InitLogger(&log.LoggerConfig{
		Level:    "info",
		Encoding: "console",
		Color:    true,
		DevMode:  true,
	}); err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}

	// Use the chi router (you can use any router you want)
	router := chi.NewRouter()
	router.Use(
		middleware.Recoverer, // Recover from panics
		middleware.RequestID, // Inject request-id
	)

	// Request logger
	var loggerConfig = logger.Config{
		Level:        zapcore.InfoLevel,
		RequestBody:  true,
		ResponseBody: true,
		IgnorePaths:  []string{},
	}
	router.Use(logger.LoggerStandardMiddleware(log.Logger.Named("server").Desugar(), loggerConfig))

	// CORS handler for REST APIs
	var corsOptions = cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodHead, http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}
	router.Use(cors.New(corsOptions).Handler)

	// Enable metrics for server
	router.Use(metrics.MetricsMiddleware(metrics.Config{}))

	// Create a server
	s, err := httpserver.New(
		httpserver.WithAddress("", "8080"),
		httpserver.WithHandler(router),
		// httpserver.WithDevCert(), // Enable a test certificate/TLS
	)
	if err != nil {
		log.Fatalf("could not create server error: %v", err)
	}

	// Start the listener and service connections.
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Errorf("Server error: %v", err)
			signal.Stop.Stop()
		}
	}()
	log.Infof("API listening on %s", s.Addr)

	// Register signal handler and wait
	signal.Stop.OnSignal(signal.DefaultStopSignals...)
	<-signal.Stop.Chan() // Wait until Stop
	signal.Stop.Wait()   // Wait until everyone cleans up
	log.Flush()          // Flush the logger

}
