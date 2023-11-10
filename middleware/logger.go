package middleware

import (
	"fmt"
	"path"
	"runtime"
	"sync"

	dt "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go"
	"github.com/kataras/golog"
)

var loggerPool sync.Pool

func init() {
	loggerPool = sync.Pool{
		New: func() interface{} {
			return &baseLogger{}
		},
	}
}

func newBaseLogger(traceName, traceID string) *baseLogger {
	logger := loggerPool.New().(*baseLogger)
	logger.traceID = traceID
	logger.traceName = traceName
	return logger
}

type baseLogger struct {
	traceID   string
	traceName string
}

// Print prints a log message without levels and colors.
func (l *baseLogger) Print(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	dt.Logger().Print(v...)
}

// Printf formats according to a format specifier and writes to `Printer#Output` without levels and colors.
func (l *baseLogger) Printf(format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	dt.Logger().Printf(format, args...)
}

// Println prints a log message without levels and colors.
// It adds a new line at the end, it overrides the `NewLine` option.
func (l *baseLogger) Println(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	dt.Logger().Println(v...)
}

// Log prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *baseLogger) Log(level golog.Level, v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	dt.Logger().Log(level, v...)
}

// Logf prints a leveled log message to the output.
// This method can be used to use custom log levels if needed.
// It adds a new line in the end.
func (l *baseLogger) Logf(level golog.Level, format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	dt.Logger().Logf(level, format, args...)
}

// Fatal `os.Exit(1)` exit no matter the level of the baseLogger.
// If the baseLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *baseLogger) Fatal(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	dt.Logger().Fatal(v...)
}

// Fatalf will `os.Exit(1)` no matter the level of the baseLogger.
// If the baseLogger's level is fatal, error, warn, info or debug
// then it will print the log message too.
func (l *baseLogger) Fatalf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	dt.Logger().Fatalf(format, args...)
}

// Error will print only when baseLogger's Level is error, warn, info or debug.
func (l *baseLogger) Error(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	dt.Logger().Error(v...)
}

// Errorf will print only when baseLogger's Level is error, warn, info or debug.
func (l *baseLogger) Errorf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	dt.Logger().Errorf(format, args...)
}

// Warn will print when baseLogger's Level is warn, info or debug.
func (l *baseLogger) Warn(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	dt.Logger().Warn(v...)
}

// Warnf will print when baseLogger's Level is warn, info or debug.
func (l *baseLogger) Warnf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	dt.Logger().Warnf(format, args...)
}

// Info will print when baseLogger's Level is info or debug.
func (l *baseLogger) Info(v ...interface{}) {
	trace := l.traceField()
	v = append(v, trace)
	dt.Logger().Info(v...)
}

// Infof will print when baseLogger's Level is info or debug.
func (l *baseLogger) Infof(format string, args ...interface{}) {
	trace := l.traceField()
	args = append(args, trace)
	dt.Logger().Infof(format, args...)
}

// Debug will print when baseLogger's Level is debug.
func (l *baseLogger) Debug(v ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	v = append(v, caller, trace)
	dt.Logger().Debug(v...)
}

// Debugf will print when baseLogger's Level is debug.
func (l *baseLogger) Debugf(format string, args ...interface{}) {
	caller := l.callerField()
	trace := l.traceField()
	args = append(args, caller, trace)
	dt.Logger().Debugf(format, args...)
}

// traceField
func (l *baseLogger) traceField() golog.Fields {
	return golog.Fields{l.traceName: l.traceID}
}

// callerField
func (l *baseLogger) callerField() golog.Fields {
	_, file, line, _ := runtime.Caller(2)
	return golog.Fields{"caller": fmt.Sprintf("%s:%d", path.Base(file), line)}
}
