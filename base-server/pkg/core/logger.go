package core

import (
	"fmt"
	"log"
	"os"
)

// simpleLogger implements Logger interface using standard library
type simpleLogger struct {
	logger *log.Logger
	fields map[string]interface{}
}

// NewLogger creates a new logger
func NewLogger() Logger {
	return &simpleLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		fields: make(map[string]interface{}),
	}
}

// Debug logs debug messages
func (l *simpleLogger) Debug(args ...interface{}) {
	l.logWithLevel("DEBUG", args...)
}

// Info logs info messages
func (l *simpleLogger) Info(args ...interface{}) {
	l.logWithLevel("INFO", args...)
}

// Warn logs warning messages
func (l *simpleLogger) Warn(args ...interface{}) {
	l.logWithLevel("WARN", args...)
}

// Error logs error messages
func (l *simpleLogger) Error(args ...interface{}) {
	l.logWithLevel("ERROR", args...)
}

// Fatal logs fatal messages and exits
func (l *simpleLogger) Fatal(args ...interface{}) {
	l.logWithLevel("FATAL", args...)
	os.Exit(1)
}

// With creates a new logger with additional fields
func (l *simpleLogger) With(key string, value interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &simpleLogger{
		logger: l.logger,
		fields: newFields,
	}
}

// logWithLevel logs with a specific level
func (l *simpleLogger) logWithLevel(level string, args ...interface{}) {
	message := fmt.Sprint(args...)

	if len(l.fields) > 0 {
		fieldsStr := ""
		for k, v := range l.fields {
			fieldsStr += fmt.Sprintf(" %s=%v", k, v)
		}
		message = fmt.Sprintf("[%s]%s %s", level, fieldsStr, message)
	} else {
		message = fmt.Sprintf("[%s] %s", level, message)
	}

	l.logger.Println(message)
}
