package tasklogger

import (
	"fmt"
	"runtime"
	"strings"

	"neo/internal/logger"

	"github.com/sirupsen/logrus"

	neopb "neo/lib/genproto/neo"
)

// 1 MB.
const maxMessageLength = 1024 * 1024

func New(exploit string, version int64, team string, sender Sender) *TaskLogger {
	return &TaskLogger{
		exploit: exploit,
		version: version,
		team:    team,
		sender:  sender,
	}
}

type TaskLogger struct {
	exploit string
	version int64
	team    string
	sender  Sender
}

func (l *TaskLogger) Debugf(format string, args ...interface{}) {
	l.logProxy(logrus.DebugLevel, 2, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "debug"))
}

func (l *TaskLogger) Infof(format string, args ...interface{}) {
	l.logProxy(logrus.InfoLevel, 2, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "info"))
}

func (l *TaskLogger) Warningf(format string, args ...interface{}) {
	l.logProxy(logrus.WarnLevel, 2, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "warning"))
}

func (l *TaskLogger) Errorf(format string, args ...interface{}) {
	l.logProxy(logrus.ErrorLevel, 2, format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "error"))
}

func (l *TaskLogger) newLine(msg, level string) *neopb.LogLine {
	return &neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: sanitizeMessage(msg),
		Level:   level,
		Team:    l.team,
	}
}

func (l *TaskLogger) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"exploit": l.exploit,
		"version": l.version,
		"team":    l.team,
	})
}

func (l *TaskLogger) logProxy(level logrus.Level, skip int, format string, args ...interface{}) {
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
