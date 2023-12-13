// badger doesn't do structured logging but is instead using Debugf and formats the strings
// with parameters. That makes analyzis a bit tricky since the same log print might differ,
// for instance startup time.
// This is a small improvement, the logs are ugly but it's possible to group and filter them
package badger

import (
	"fmt"
	"log/slog"
	"strings"
)

type Logger struct {
	l *slog.Logger
}

func NewLogger() *Logger {
	l := slog.Default().With(slog.String("logSubSystem", "badgerLogger"))
	return &Logger{l}
}

func (l *Logger) Errorf(f string, v ...interface{}) {
	m, args := format(f, v...)
	l.l.Error(m, args)
}

func (l *Logger) Warningf(f string, v ...interface{}) {
	m, args := format(f, v...)
	l.l.Warn(m, args)
}

func (l *Logger) Infof(f string, v ...any) {
	m, args := format(f, v...)
	l.l.Info(m, args)
}

func (l *Logger) Debugf(f string, v ...interface{}) {
	m, args := format(f, v...)
	l.l.Debug(m, args)
}

func format(f string, v ...any) (string, slog.Attr) {
	attr := make([]any, 0, len(v))
	for i, val := range v {
		attr = append(attr, slog.Any(fmt.Sprintf("arg%d", i), val))
	}
	group := slog.Group("formatArgs", attr...)
	f = strings.ReplaceAll(f, "\n", "") // no newlines
	return f, group
}
