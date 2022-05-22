package tasklogger

import (
	"fmt"

	"github.com/sirupsen/logrus"

	neopb "neo/lib/genproto/neo"
)

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

func (l TaskLogger) Debugf(format string, args ...interface{}) {
	l.getLogger().Debugf(format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "debug"))
}

func (l TaskLogger) Infof(format string, args ...interface{}) {
	l.getLogger().Infof(format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "info"))
}

func (l TaskLogger) Warningf(format string, args ...interface{}) {
	l.getLogger().Warningf(format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "warning"))
}

func (l TaskLogger) Errorf(format string, args ...interface{}) {
	l.getLogger().Errorf(format, args...)
	msg := fmt.Sprintf(format, args...)
	l.sender.Add(l.newLine(msg, "error"))
}

func (l TaskLogger) newLine(msg, level string) *neopb.LogLine {
	return &neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: msg,
		Level:   level,
		Team:    l.team,
	}
}

func (l TaskLogger) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"exploit": l.exploit,
		"version": l.version,
		"team":    l.team,
	})
}
