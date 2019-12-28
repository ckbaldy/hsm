package logger

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Log provides a logrus Logger
var Log = logrus.New()

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.ForceColors = true
	formatter.ForceFormatting = true
	formatter.FullTimestamp = true
	Log.Formatter = formatter
	//Log.Level = logrus.DebugLevel
}
