package logging

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Logger wraps the logrus logger
type Logger struct {
	*logrus.Logger
}
type contextKey string

const loggerKey contextKey = "logger"

var defLogger = NewLogger("info")

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	return defLogger
}

// NewLogger creates a new logger.
func NewLogger(level string) *Logger {
	//once.Do(func() {
	logrusLogger := logrus.New()
	logrusLogger.SetReportCaller(true)
	logrusLogger.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}
	logrusLogger.SetOutput(os.Stdout)
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrusLogger.Error(err)
		logrusLevel = logrus.InfoLevel
	}
	logrusLogger.SetLevel(logrusLevel)

	return &Logger{logrusLogger}
	//})
}
