package tasklogger

import (
	"fmt"

	"github.com/sirupsen/logrus"

	neopb "neo/lib/genproto/neo"
)

func New(exploit string, version int64, team string, sender *Sender) *TaskLogger {
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
	sender  *Sender
}

func (l TaskLogger) Debugf(format string, args ...interface{}) {
	l.getLogger().Debugf(format, args...)
	msg := fmt.Sprintf(format, args...)
	line := neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: msg,
		Level:   "debug",
		Team:    l.team,
	}
	l.sender.Add(&line)
}

func (l TaskLogger) Infof(format string, args ...interface{}) {
	l.getLogger().Infof(format, args...)
	msg := fmt.Sprintf(format, args...)
	line := neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: msg,
		Level:   "info",
		Team:    l.team,
	}
	l.sender.Add(&line)
}

func (l TaskLogger) Warningf(format string, args ...interface{}) {
	l.getLogger().Warningf(format, args...)
	msg := fmt.Sprintf(format, args...)
	line := neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: msg,
		Level:   "warning",
		Team:    l.team,
	}
	l.sender.Add(&line)
}

func (l TaskLogger) Errorf(format string, args ...interface{}) {
	l.getLogger().Errorf(format, args...)
	msg := fmt.Sprintf(format, args...)
	line := neopb.LogLine{
		Exploit: l.exploit,
		Version: l.version,
		Message: msg,
		Level:   "error",
		Team:    l.team,
	}
	l.sender.Add(&line)
}

func (l TaskLogger) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"exploit": l.exploit,
		"version": l.version,
		"team":    l.team,
	})
}
