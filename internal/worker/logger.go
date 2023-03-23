package worker

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type WorkerLogger struct{}

func NewWorkerLogger() *WorkerLogger {
	return &WorkerLogger{}
}

func (l *WorkerLogger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

func (l *WorkerLogger) Debug(args ...interface{}) {
	l.Print(zerolog.DebugLevel, args...)
}

func (l *WorkerLogger) Info(args ...interface{}) {
	l.Print(zerolog.InfoLevel, args...)
}

func (l *WorkerLogger) Warn(args ...interface{}) {
	l.Print(zerolog.WarnLevel, args...)
}

func (l *WorkerLogger) Error(args ...interface{}) {
	l.Print(zerolog.ErrorLevel, args...)
}

func (l *WorkerLogger) Fatal(args ...interface{}) {
	l.Print(zerolog.FatalLevel, args...)
}
