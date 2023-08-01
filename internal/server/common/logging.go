package common

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoggingServer struct {
	Logger *logrus.Entry
}

func NewLoggingServer(serverName string) LoggingServer {
	return LoggingServer{Logger: logrus.WithField("server", serverName)}
}

func (s *LoggingServer) WrapErrorf(ctx context.Context, code codes.Code, fmt string, values ...any) error {
	err := status.Errorf(code, fmt, values...)
	s.GetMethodLogger(ctx).Errorf("%v", err)
	return err // nolint
}

func (s *LoggingServer) LogRequest(ctx context.Context, r any) {
	s.GetMethodLogger(ctx).Infof("Request: %v", r)
}

func (s *LoggingServer) GetMethodLogger(ctx context.Context) *logrus.Entry {
	if method, ok := grpc.Method(ctx); ok {
		return s.Logger.WithField("method", method)
	}
	return s.Logger
}
