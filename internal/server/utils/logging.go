package utils

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WrapErrorf(code codes.Code, fmt string, values ...interface{}) error {
	err := status.Errorf(code, fmt, values...)
	logrus.Errorf("%v", err)
	return err // nolint
}
