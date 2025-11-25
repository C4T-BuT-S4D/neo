package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the global zap logger with colored, human-readable console output.
func Init(debug bool) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}

	// Check environment variable for log level override.
	if ll := os.Getenv("NEO_LOG_LEVEL"); ll != "" {
		if err := level.UnmarshalText([]byte(ll)); err != nil {
			// Fallback to info if parsing fails.
			level = zapcore.InfoLevel
		}
	}

	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
	}
	config.EncodeCaller = zapcore.ShortCallerEncoder
	config.ConsoleSeparator = " "

	encoder := zapcore.NewConsoleEncoder(config)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stderr),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	zap.ReplaceGlobals(logger)
}

// L returns the global logger.
func L() *zap.Logger {
	return zap.L()
}

// S returns the global sugared logger.
func S() *zap.SugaredLogger {
	return zap.S()
}

// Sync flushes any buffered log entries.
func Sync() {
	_ = zap.L().Sync()
}

