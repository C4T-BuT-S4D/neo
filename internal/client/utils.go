package client

import (
	"context"
	"errors"
	"io"

	"go.uber.org/zap"
)

func checkStreamError(tp string, err, streamErr error) bool {
	if errors.Is(err, io.EOF) {
		zap.L().Error("Stream closed", zap.String("type", tp))
		return false
	}
	if errors.Is(streamErr, context.Canceled) {
		zap.L().Debug("Context cancelled", zap.String("type", tp))
		return false
	}
	if err != nil {
		zap.L().Error("Error reading from stream", zap.String("type", tp), zap.Error(err))
		return false
	}
	return true
}
