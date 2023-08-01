package joblogger

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/c4t-but-s4d/neo/internal/logger"
	logspb "github.com/c4t-but-s4d/neo/proto/go/logs"

	"github.com/sirupsen/logrus"
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

func (l *JobLogger) Debugf(format string, args ...interface{}) {
	l.logProxy(logrus.DebugLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "debug"))
}

func (l *JobLogger) Infof(format string, args ...interface{}) {
	l.logProxy(logrus.InfoLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "info"))
}

func (l *JobLogger) Warningf(format string, args ...interface{}) {
	l.logProxy(logrus.WarnLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "warning"))
}

func (l *JobLogger) Errorf(format string, args ...interface{}) {
	l.logProxy(logrus.ErrorLevel, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "error"))
}

func (l *JobLogger) newLine(msg, level string) *logspb.LogLine {
	return &logspb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: sanitizeMessage(msg),
		Level:   level,
		Team:    l.team,
	}
}

func (l *JobLogger) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"exploit": l.exploit,
		"version": l.version,
		"team":    l.team,
	})
}

func (l *JobLogger) logProxy(level logrus.Level, format string, args ...interface{}) {
	if logrus.IsLevelEnabled(level) {
		l.
			getLogger().
			WithField(logger.CustomKeyFile, fileInfo(3)).
			Logf(level, format, args...)
	}
}

func sanitizeMessage(msg string) string {
	if len(msg) > maxMessageLength {
		msg = msg[:maxMessageLength]
	}
	return msg
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}
