// badger doesn't do structured logging but is instead using xxf and formats the strings
// with parameters. That makes analyzis a bit tricky since the same error might
package logging

import (
	"fmt"
	"log/slog"
)

type BadgerLogger struct {
	l *slog.Logger
}

func New() *BadgerLogger {
	l := slog.Default().With(slog.String("logSubSystem", "badgerLogger"))

	return &BadgerLogger{l}
}

func (l *BadgerLogger) Errorf(f string, v ...interface{}) {
	l.l.Error(fmt.Sprintf(f, v...))
}

func (l *BadgerLogger) Warningf(f string, v ...interface{}) {
	l.l.Warn(fmt.Sprintf(f, v...))
}

func (l *BadgerLogger) Infof(f string, v ...interface{}) {
	l.l.Info(fmt.Sprintf(f, v...))
}

func (l *BadgerLogger) Debugf(f string, v ...interface{}) {
	l.l.Debug(fmt.Sprintf(f, v...))
}
