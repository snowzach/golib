package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snowzach/golib/conf"
)

func main() {

	// Our defaults
	defaults := map[string]interface{}{
		// Metrics endpoint
		"metrics.enabled": true,

		// Logger Defaults
		"logger.level": "info",
		"logger.color": false,
		"logger.name":  "default",

		"server.host":     "0.0.0.0",
		"server.base_url": "/base/url",
	}

	// Load some settings from a config file
	var cfgFile string = "example.yaml"

	// Set an environment variable to show how that works.
	os.Setenv("LOGGER_NAME", "somevalue")

	// Parse the defaults, load a config file and finally environment.
	// Put everything into the global conf instance.
	if err := conf.C.Parse(
		conf.WithMap(defaults),
		conf.WithFile(cfgFile),
		conf.WithEnv(),
	); err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	// Get by name as a bool as a simple example.
	fmt.Printf("metrics.enabled: %v\n", conf.C.Bool("metrics.enabled"))

	// Unmarshal to a struct for speed
	var loggerConfig struct {
		Level string `conf:"level"`
		Color bool   `conf:"color"`
		Name  string `conf:"name"`
	}
	if err := conf.C.UnmarshalWithOpts(&loggerConfig,
		conf.WithPath("logger"),
		conf.WithTag("conf"),
	); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("logger: %v", toJSON(loggerConfig))

	///////////////////////////////////////////////////////////////////////
	// Also supports setting defaults in the struct but also allows from
	// config and environment.

	os.Setenv("SERVER_DEBUG", "true")
	os.Setenv("SERVER_OPTIONS", "x y z")
	os.Setenv("SERVER_BASE_URL", "/base/url/custom")

	var serverConfig struct {
		Host    string   `conf:"host"`                                  // Default set in the global defaults
		Port    int      `conf:"port" default:"8080"`                   // Default set using struct tag
		Debug   bool     `conf:"debug" default:"false"`                 // Default set using struct tag, override with env
		Headers []string `conf:"headers" default:"[\"a\",\"b\",\"c\"]"` // Struct tags use json for defaults
		Options []string `conf:"options"`
		BaseURL string   `conf:"base_url"`
	}
	if err := conf.C.UnmarshalWithOpts(&serverConfig,
		conf.WithPath("server"),
		conf.WithTag("conf"),
		conf.WithStructDefaults(true),
		conf.WithStructEnvironment(true),
	); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("server: %s", toJSON(serverConfig))

	///////////////////////////////////////////////////////////////////////
	// If you just want to use for environment variables with defaults and config or files, it supports that also.

	os.Setenv("DB_HOST", "somehost.domain.com")

	var dbConfig struct {
		Host     string `conf:"host" default:"localhost"`
		Port     int    `conf:"port" default:"5432"`
		Username string `conf:"username" default:"postgres"`
		Password string `conf:"password" default:"password"`
	}
	if err := conf.UnmarshalEnvStruct("db", &dbConfig, "conf"); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("db: %s", toJSON(dbConfig))

}

func toJSON(in interface{}) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(in)
	return buf.String()
}
