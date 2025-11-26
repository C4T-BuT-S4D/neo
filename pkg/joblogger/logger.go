package joblogger

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

// 1 MB.
const maxMessageLength = 1024 * 1024

func New(exploit string, version int64, team string, sender Sender) *JobLogger {
	logger := zap.L().
		WithOptions(zap.AddCallerSkip(1)).
		With(
			zap.String("exploit", exploit),
			zap.Int64("version", version),
			zap.String("team", team),
		)
	return &JobLogger{
		exploit: exploit,
		version: version,
		team:    team,
		sender:  sender,
		logger:  logger,
	}
}

type JobLogger struct {
	exploit string
	version int64
	team    string
	sender  Sender
	logger  *zap.Logger
}

func (l *JobLogger) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Debug(msg)
	l.sender.Add(l.newLine(msg, "debug"))
}

func (l *JobLogger) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Info(msg)
	l.sender.Add(l.newLine(msg, "info"))
}

func (l *JobLogger) Warningf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Warn(msg)
	l.sender.Add(l.newLine(msg, "warning"))
}

func (l *JobLogger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Error(msg)
	l.sender.Add(l.newLine(msg, "error"))
}

func (l *JobLogger) newLine(msg, level string) *logspb.LogLine {
	return &logspb.LogLine{
		Exploit:   l.exploit,
		Version:   l.version,
		Message:   sanitizeMessage(msg),
		Level:     level,
		Team:      l.team,
		Timestamp: timestamppb.Now(),
	}
}

func sanitizeMessage(msg string) string {
	if len(msg) > maxMessageLength {
		msg = msg[:maxMessageLength]
	}
	return msg
}
