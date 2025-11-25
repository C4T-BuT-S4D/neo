package logging

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	Logger *zap.Logger
}

func NewServer(serverName string) Server {
	return Server{Logger: zap.L().Named(serverName)}
}

func (s *Server) WrapErrorf(ctx context.Context, code codes.Code, fmt string, values ...any) error {
	err := status.Errorf(code, fmt, values...)
	s.GetMethodLogger(ctx).Error(err.Error())
	return err
}

func (s *Server) LogRequest(ctx context.Context, r any) {
	s.GetMethodLogger(ctx).Info("Request", zap.Any("payload", r))
}

func (s *Server) GetMethodLogger(ctx context.Context) *zap.Logger {
	if method, ok := grpc.Method(ctx); ok {
		return s.Logger.With(zap.String("method", method))
	}
	return s.Logger
}
