package gommonlog

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/labstack/gommon/log"
)

//
// =========================
// Logger
// =========================
//

type Logger struct {
	common *log.Logger // DEBUG, INFO, WARN
	error  *log.Logger // ERROR, FATAL, PANIC
}

func New() *Logger {
	common := log.New("")
	common.SetLevel(log.DEBUG)
	common.SetHeader(`${time_rfc3339} ${level}${message}`)

	errLogger := log.New("")
	errLogger.SetLevel(log.ERROR)
	errLogger.SetHeader(`${time_rfc3339} ${level}${message}`)

	return &Logger{
		common: common,
		error:  errLogger,
	}
}

//
// =========================
// caller helper
// =========================
//

func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}

	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]

	return fmt.Sprintf("%s:%d", file, line)
}

//
// =========================
// internal helpers
// =========================
//

func (l *Logger) errorWithSkip(skip int, msg string) {
	l.error.Error(msg + " " + caller(skip))
}

func formatFields(msg string, fields map[string]interface{}) string {
	if len(fields) == 0 {
		return msg
	}

	parts := make([]string, 0, len(fields))
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	if msg == "" {
		return strings.Join(parts, " ")
	}

	return msg + " | " + strings.Join(parts, " ")
}

func formatErr(msg string, err error) string {
	if err == nil {
		return msg
	}
	return msg + " | err=" + err.Error()
}

//
// =========================
// base methods
// =========================
//

func (l *Logger) Debug(args ...interface{}) {
	l.common.Debug(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.common.Debugf(format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.common.Info(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.common.Infof(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.common.Warn(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.common.Warnf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.errorWithSkip(3, msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.errorWithSkip(3, msg)
}

func (l *Logger) Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.errorWithSkip(3, msg)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.errorWithSkip(3, msg)
	os.Exit(1)
}

func (l *Logger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.errorWithSkip(3, msg)
	panic(msg)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.errorWithSkip(3, msg)
	panic(msg)
}

//
// =========================
// WithField / WithFields
// =========================
//

type FieldLogger struct {
	parent *Logger
	fields map[string]interface{}
}

func (l *Logger) WithField(key string, value interface{}) *FieldLogger {
	return &FieldLogger{
		parent: l,
		fields: map[string]interface{}{
			key: value,
		},
	}
}

func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		parent: l,
		fields: fields,
	}
}

func (f *FieldLogger) WithField(key string, value interface{}) *FieldLogger {
	f.fields[key] = value
	return f
}

func (f *FieldLogger) Debug(args ...interface{}) {
	f.parent.Debug(formatFields(fmt.Sprint(args...), f.fields))
}

func (f *FieldLogger) Debugf(format string, args ...interface{}) {
	f.parent.Debug(formatFields(fmt.Sprintf(format, args...), f.fields))
}

func (f *FieldLogger) Info(args ...interface{}) {
	f.parent.Info(formatFields(fmt.Sprint(args...), f.fields))
}

func (f *FieldLogger) Infof(format string, args ...interface{}) {
	f.parent.Infof(formatFields(fmt.Sprintf(format, args...), f.fields))
}

func (f *FieldLogger) Warn(args ...interface{}) {
	f.parent.Warn(formatFields(fmt.Sprint(args...), f.fields))
}

func (f *FieldLogger) Warnf(format string, args ...interface{}) {
	f.parent.Warnf(formatFields(fmt.Sprintf(format, args...), f.fields))
}

func (f *FieldLogger) Error(args ...interface{}) {
	msg := formatFields(fmt.Sprint(args...), f.fields)
	f.parent.errorWithSkip(3, msg)
}

func (f *FieldLogger) Errorf(format string, args ...interface{}) {
	msg := formatFields(fmt.Sprintf(format, args...), f.fields)
	f.parent.errorWithSkip(3, msg)
}

func (f *FieldLogger) Fatal(args ...interface{}) {
	msg := formatFields(fmt.Sprint(args...), f.fields)
	f.parent.errorWithSkip(3, msg)
	os.Exit(1)
}

func (f *FieldLogger) Fatalf(format string, args ...interface{}) {
	msg := formatFields(fmt.Sprintf(format, args...), f.fields)
	f.parent.errorWithSkip(3, msg)
	os.Exit(1)
}

func (f *FieldLogger) Panic(args ...interface{}) {
	msg := formatFields(fmt.Sprint(args...), f.fields)
	f.parent.errorWithSkip(3, msg)
	panic(msg)
}

func (f *FieldLogger) Panicf(format string, args ...interface{}) {
	msg := formatFields(fmt.Sprintf(format, args...), f.fields)
	f.parent.errorWithSkip(3, msg)
	panic(msg)
}

//
// =========================
// WithError
// =========================
//

type ErrorLogger struct {
	parent *Logger
	err    error
}

func (l *Logger) WithError(err error) *ErrorLogger {
	return &ErrorLogger{
		parent: l,
		err:    err,
	}
}

func (f *FieldLogger) WithError(err error) *ErrorLogger {
	msg := formatFields("", f.fields)
	if msg != "" {
		err = fmt.Errorf("%s | %w", msg, err)
	}

	return &ErrorLogger{
		parent: f.parent,
		err:    err,
	}
}

//
// =========================
// WithError methods
// =========================
//

func (e *ErrorLogger) Debug(args ...interface{}) {
	e.parent.Debug(formatErr(fmt.Sprint(args...), e.err))
}

func (e *ErrorLogger) Debugf(format string, args ...interface{}) {
	e.parent.Debug(formatErr(fmt.Sprintf(format, args...), e.err))
}

func (e *ErrorLogger) Info(args ...interface{}) {
	e.parent.Info(formatErr(fmt.Sprint(args...), e.err))
}

func (e *ErrorLogger) Infof(format string, args ...interface{}) {
	e.parent.Infof(formatErr(fmt.Sprintf(format, args...), e.err))
}

func (e *ErrorLogger) Warn(args ...interface{}) {
	e.parent.Warn(formatErr(fmt.Sprint(args...), e.err))
}

func (e *ErrorLogger) Warnf(format string, args ...interface{}) {
	e.parent.Warnf(formatErr(fmt.Sprintf(format, args...), e.err))
}

func (e *ErrorLogger) Error(args ...interface{}) {
	msg := formatErr(fmt.Sprint(args...), e.err)
	e.parent.errorWithSkip(3, msg)
}

func (e *ErrorLogger) Errorf(format string, args ...interface{}) {
	msg := formatErr(fmt.Sprintf(format, args...), e.err)
	e.parent.errorWithSkip(3, msg)
}

func (e *ErrorLogger) Fatal(args ...interface{}) {
	msg := formatErr(fmt.Sprint(args...), e.err)
	e.parent.errorWithSkip(3, msg)
	os.Exit(1)
}

func (e *ErrorLogger) Fatalf(format string, args ...interface{}) {
	msg := formatErr(fmt.Sprintf(format, args...), e.err)
	e.parent.errorWithSkip(3, msg)
	os.Exit(1)
}

func (e *ErrorLogger) Panic(args ...interface{}) {
	msg := formatErr(fmt.Sprint(args...), e.err)
	e.parent.errorWithSkip(3, msg)
	panic(msg)
}

func (e *ErrorLogger) Panicf(format string, args ...interface{}) {
	msg := formatErr(fmt.Sprintf(format, args...), e.err)
	e.parent.errorWithSkip(3, msg)
	panic(msg)
}
