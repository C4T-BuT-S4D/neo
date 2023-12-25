package mu

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"nhooyr.io/websocket"
)

func NewHandler(grpcServer *grpc.Server, opts ...Option) http.Handler {
	cfg := &Config{
		http2Server: &http2.Server{},
		httpHandler: http.DefaultServeMux,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return h2c.NewHandler(
		&Server{
			grpcServer: grpcServer,
			webServer: grpcweb.WrapServer(
				grpcServer,
				grpcweb.WithWebsockets(true),
				grpcweb.WithOriginFunc(func(string) bool {
					return true
				}),
				grpcweb.WithWebsocketOriginFunc(func(*http.Request) bool {
					return true
				}),
				grpcweb.WithWebsocketCompressionMode(websocket.CompressionDisabled),
			),
			httpHandler: cfg.httpHandler,
		},
		cfg.http2Server,
	)
}

type Server struct {
	grpcServer  *grpc.Server
	webServer   *grpcweb.WrappedGrpcServer
	httpHandler http.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc":
		s.grpcServer.ServeHTTP(w, r)
	case s.webServer.IsAcceptableGrpcCorsRequest(r) || s.webServer.IsGrpcWebRequest(r) || s.webServer.IsGrpcWebSocketRequest(r):
		s.webServer.ServeHTTP(w, r)
	default:
		s.httpHandler.ServeHTTP(w, r)
	}
}
