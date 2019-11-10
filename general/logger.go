package general

import (
	"github.com/kataras/golog"
)

// Logger .
type Logger interface {
	Print(v ...interface{})
	Printf(format string, args ...interface{})
	Println(v ...interface{})
	Log(level golog.Level, v ...interface{})
	Logf(level golog.Level, format string, args ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, args ...interface{})
	Error(v ...interface{})
	Errorf(format string, args ...interface{})
	Warn(v ...interface{})
	Warnf(format string, args ...interface{})
	Info(v ...interface{})
	Infof(format string, args ...interface{})
	Debug(v ...interface{})
	Debugf(format string, args ...interface{})
}
