package logging

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	Logger *logrus.Entry
}

func NewServer(serverName string) Server {
	return Server{Logger: logrus.WithField("server", serverName)}
}

func (s *Server) WrapErrorf(ctx context.Context, code codes.Code, fmt string, values ...any) error {
	err := status.Errorf(code, fmt, values...)
	s.GetMethodLogger(ctx).Errorf("%v", err)
	return err
}

func (s *Server) LogRequest(ctx context.Context, r any) {
	s.GetMethodLogger(ctx).Infof("Request: %v", r)
}

func (s *Server) GetMethodLogger(ctx context.Context) *logrus.Entry {
	if method, ok := grpc.Method(ctx); ok {
		return s.Logger.WithField("method", method)
	}
	return s.Logger
}

func WrapErrorf(code codes.Code, fmt string, values ...any) error {
	err := status.Errorf(code, fmt, values...)
	logrus.Errorf("%v", err)
	return err
}
