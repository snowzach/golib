package httpserver

import "net/http"

// WithConfig applies the whole config
func WithConfig(config *Config) Option {
	return func(c *Config) {
		*c = *config
	}
}

// WithAddress sets the listener address
func WithAddress(host string, port string) Option {
	return func(c *Config) {
		c.Host = host
		c.Port = port
	}
}

// WithDevCert enables a development tls certificate.
// NEVER USE THIS IN PRODUCTION
func WithDevCert() Option {
	return func(c *Config) {
		c.TLS = true
		c.DevCert = true
	}
}

// WithCertFiles sets the certFile and keyFile for tls
func WithCertFiles(certFile, keyFile string) Option {
	return func(c *Config) {
		c.TLS = true
		c.CertFile = certFile
		c.KeyFile = keyFile
	}
}

// WithHandler sets the http handler/router
func WithHandler(handler http.Handler) Option {
	return func(c *Config) {
		c.Handler = handler
	}
}
