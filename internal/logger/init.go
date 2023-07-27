// Confidential. Intellectual property of Pinely Holdings Pte. Ltd. Refer to CONFIDENTIAL file in the root for details

package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

var once = sync.Once{}

func Init() {
	once.Do(func() {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:            true,
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02T15:04:05.000Z07:00",
			DisableLevelTruncation: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return "", fmt.Sprintf(" %s:%d", filename, f.Line)
			},
		})
		logrus.SetReportCaller(true)
	})
}
