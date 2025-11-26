package logging

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var initOnce sync.Once

type CheckedCloser interface {
	Close()
}

type checkedCloserImpl func()

func (c checkedCloserImpl) Close() {
	c()
}

// Init initializes the global zap logger with colored, human-readable console output.
func Init(level string) CheckedCloser {
	initOnce.Do(func() {
		lvl := parseLogLevel(level, zap.InfoLevel)
		zap.ReplaceGlobals(createLogger(lvl))
	})

	return checkedCloserImpl(syncLogger)
}

func parseLogLevel(levelStr string, defaultLevel zapcore.Level) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	case "panic":
		return zap.PanicLevel
	default:
		return defaultLevel
	}
}

func SetLogLevel(level string) {
	lvl := parseLogLevel(level, zap.InfoLevel)
	syncLogger()
	zap.ReplaceGlobals(createLogger(lvl))
}

func syncLogger() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.EINVAL) {
		fmt.Printf("failed to sync logger: %v\n", err)
	}
}

func createLogger(level zapcore.Level) *zap.Logger {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(config)

	stderr := zapcore.Lock(os.Stderr)
	core := zapcore.NewCore(encoder, stderr, level)
	return zap.New(core, zap.WithCaller(true))
}
