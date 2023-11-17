package mu

import (
	"net/http"

	"golang.org/x/net/http2"
)

type Config struct {
	http2Server *http2.Server
	httpHandler http.Handler
}

type Option func(cfg *Config)

func WithHTTP2Server(server *http2.Server) Option {
	return func(cfg *Config) {
		cfg.http2Server = server
	}
}

func WithHTTPHandler(handler http.Handler) Option {
	return func(cfg *Config) {
		cfg.httpHandler = handler
	}
}
