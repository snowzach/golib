package httpserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/snowzach/certtools"
	"github.com/snowzach/certtools/autocert"
)

type Config struct {
	Host     string `conf:"host"`
	Port     string `conf:"port" default:"8080"`
	TLS      bool   `conf:"tls"`
	DevCert  bool   `conf:"devcert"`
	CertFile string `conf:"certfile"`
	KeyFile  string `conf:"keyfile"`
	Handler  http.Handler
}

type Option func(c *Config)

type Server struct {
	config    *Config
	tlsConfig *tls.Config
	*http.Server
}

// New will setup the API listener. You can specify a configuration
// or you can set config to nil and provide a list of options.
func New(opts ...Option) (*Server, error) {

	// If config not specified, create a new one and apply options
	var config = &Config{}
	for _, opt := range opts {
		opt(config)
	}

	// Setup server
	s := &Server{
		config: config,
		Server: &http.Server{
			Addr:    net.JoinHostPort(config.Host, config.Port),
			Handler: config.Handler,
		},
	}

	// Configure TLS
	if config.TLS {
		var (
			err  error
			cert tls.Certificate
		)
		if s.config.DevCert {
			cert, err = autocert.New(autocert.InsecureStringReader("localhost"))
			if err != nil {
				return nil, fmt.Errorf("could not generate autocert server certificate: %w", err)
			}
			// Horrible unstoppable disclaimer
			fmt.Println("*** GENERATING A DEV CERTIFICATE - THIS SHOULD NEVER BE USED IN PRODUCTION ***")
		} else {
			// Load keys from file
			cert, err = tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("could not load server certificate: %w", err)
			}
		}
		// Sane/Safe defaults
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   certtools.SecureTLSMinVersion(),
			CipherSuites: certtools.SecureTLSCipherSuites(),
		}
	}

	return s, nil
}

// ListenAndServe will listen for requests
func (s *Server) ListenAndServe() error {

	// Listen
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %w", s.Addr, err)
	}

	// Enable TLS?
	if s.config.TLS {
		// Wrap the listener in a TLS Listener
		listener = tls.NewListener(listener, s.TLSConfig)
	}

	return s.Serve(listener)

}
