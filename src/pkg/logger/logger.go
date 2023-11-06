package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Logger struct {
	prefixers []GetPrefixesFunc
}

func New() *Logger {
	return &Logger{prefixers: []GetPrefixesFunc{}}
}

func (l *Logger) Debugf(ctx context.Context, message string, args ...interface{}) {
	l.msg(ctx, "debug", fmt.Sprintf(message, args...))
}

func (l *Logger) Debug(ctx context.Context, message string) {
	l.msg(ctx, "debug", message)
}

func (l *Logger) Infof(ctx context.Context, message string, args ...interface{}) {
	l.msg(ctx, "info", fmt.Sprintf(message, args...))
}

func (l *Logger) Info(ctx context.Context, message string) {
	l.msg(ctx, "info", message)
}

func (l *Logger) Warnf(ctx context.Context, message string, args ...interface{}) {
	l.msg(ctx, "warn", fmt.Sprintf(message, args...))
}

func (l *Logger) Warn(ctx context.Context, message string) {
	l.msg(ctx, "warn", message)
}

func (l *Logger) Errorf(ctx context.Context, message string, args ...interface{}) {
	l.msg(ctx, "error", fmt.Sprintf(message, args...))
}

func (l *Logger) Error(ctx context.Context, message string) {
	l.msg(ctx, "error", message)
}

func (l *Logger) Messsage(ctx context.Context, level string, message string) {
	l.msg(ctx, level, message)
}

func (l *Logger) Err(ctx context.Context, err error) {
	l.msg(ctx, "error", err.Error())
}

func (l *Logger) Fatalf(ctx context.Context, err error, args ...interface{}) {
	l.msg(ctx, "fatal", fmt.Sprintf(err.Error(), args...))
	os.Exit(1)
}

func (l *Logger) Fatal(ctx context.Context, err error) {
	l.msg(ctx, "fatal", err.Error())
	os.Exit(1)
}

func (l *Logger) msg(ctx context.Context, level string, fullText string) {
	ctxPrefixString := l.GetPrefixForContext(ctx)

	prefixString := fmt.Sprintf("%s [%+5s]%s ", time.Now().UTC().Format("2006-01-02 15:04:05"), strings.ToUpper(level), ctxPrefixString)

	fullText = prefixString + strings.ReplaceAll(fullText, "\n", "\n"+prefixString)

	//nolint:forbidigo
	fmt.Println(fullText)
}
