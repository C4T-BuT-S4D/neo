package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

var once = sync.Once{}

func Init() {
	once.Do(func() {
		logrus.SetFormatter(&CustomFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02T15:04:05.000Z07:00",
			DisableLevelTruncation: false,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return "", fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
		logrus.SetReportCaller(true)
	})
	if ll := os.Getenv("NEO_LOG_LEVEL"); ll != "" {
		level, err := logrus.ParseLevel(ll)
		if err != nil {
			logrus.Fatalf("Failed to parse log level %v: %v", ll, err)
		}
		logrus.SetLevel(level)
	}
}
