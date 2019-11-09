package general

import (
	"fmt"
	"strings"

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

func newFreedomLogger(traceID string) *freedomLogger {
	return &freedomLogger{
		ctx: []interface{}{traceID},
	}
}

type freedomLogger struct {
	ctx []interface{}
}

// Print prints a log message without levels and colors.
func (l *freedomLogger) Print(v ...interface{}) {
	globalApp.IrisApp.Logger().Print(l.format(v...))
}

// Printf formats according to a format specifier and writes to `Printer#Output` without levels and colors.
func (l *freedomLogger) Printf(format string, args ...interface{}) {
	l.Print(fmt.Sprintf(format, args...))
}

// Println prints a log message without levels and colors.
// It adds a new line at the end, it overrides the `NewLine` option.
func (l *freedomLogger) Println(v ...interface{}) {
	globalApp.IrisApp.Logger().Println(l.format(v...))
}

// Log prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *freedomLogger) Log(level golog.Level, v ...interface{}) {
	globalApp.IrisApp.Logger().Log(level, l.format(v...))
}

// Logf prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *freedomLogger) Logf(level golog.Level, format string, args ...interface{}) {
	l.Log(level, fmt.Sprintf(format, args...))
}

// Fatal `os.Exit(1)` exit no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *freedomLogger) Fatal(v ...interface{}) {
	globalApp.IrisApp.Logger().Fatal(l.format(v...))
}

// Fatalf will `os.Exit(1)` no matter the level of the freedomLogger.
// If the freedomLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *freedomLogger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg)
}

// Error will print only when freedomLogger's Level is error, warn, info or debug.
func (l *freedomLogger) Error(v ...interface{}) {
	globalApp.IrisApp.Logger().Error(l.format(v...))
}

// Errorf will print only when freedomLogger's Level is error, warn, info or debug.
func (l *freedomLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
}

// Warn will print when freedomLogger's Level is warn, info or debug.
func (l *freedomLogger) Warn(v ...interface{}) {
	globalApp.IrisApp.Logger().Warn(l.format(v...))
}

// Warnf will print when freedomLogger's Level is warn, info or debug.
func (l *freedomLogger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Warn(msg)
}

// Info will print when freedomLogger's Level is info or debug.
func (l *freedomLogger) Info(v ...interface{}) {
	globalApp.IrisApp.Logger().Info(l.format(v...))
}

// Infof will print when freedomLogger's Level is info or debug.
func (l *freedomLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg)
}

// Debug will print when freedomLogger's Level is debug.
func (l *freedomLogger) Debug(v ...interface{}) {
	globalApp.IrisApp.Logger().Debug(l.format(v...))
}

// Debugf will print when freedomLogger's Level is debug.
func (l *freedomLogger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg)
}

// format
func (l *freedomLogger) format(v ...interface{}) string {
	var list []string
	for _, i := range append(l.ctx, v...) {
		list = append(list, fmt.Sprint(i))
	}

	return strings.Join(list, " ")
}
