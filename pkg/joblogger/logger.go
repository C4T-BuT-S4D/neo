package joblogger

import (
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

// 1 MB.
const maxMessageLength = 1024 * 1024

func New(exploit string, version int64, team string, sender Sender) *JobLogger {
	return &JobLogger{
		exploit: exploit,
		version: version,
		team:    team,
		sender:  sender,
	}
}

type JobLogger struct {
	exploit string
	version int64
	team    string
	sender  Sender
}

func (l *JobLogger) Debugf(format string, args ...any) {
	l.logProxy(zapcore.DebugLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "debug"))
}

func (l *JobLogger) Infof(format string, args ...any) {
	l.logProxy(zapcore.InfoLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "info"))
}

func (l *JobLogger) Warningf(format string, args ...any) {
	l.logProxy(zapcore.WarnLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "warning"))
}

func (l *JobLogger) Errorf(format string, args ...any) {
	l.logProxy(zapcore.ErrorLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
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

func (l *JobLogger) getLogger() *zap.Logger {
	return zap.L().With(
		zap.String("exploit", l.exploit),
		zap.Int64("version", l.version),
		zap.String("team", l.team),
	)
}

func (l *JobLogger) logProxy(level zapcore.Level, format string, args ...any) {
	if ce := l.getLogger().Check(level, fmt.Sprintf(format, args...)); ce != nil {
		ce.Caller = zapcore.NewEntryCaller(fileInfo(3))
		ce.Write()
	}
}

func sanitizeMessage(msg string) string {
	if len(msg) > maxMessageLength {
		msg = msg[:maxMessageLength]
	}
	return msg
}

func fileInfo(skip int) (uintptr, string, int, bool) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return 0, "<???>", 1, false
	}
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	return pc, file, line, true
}
