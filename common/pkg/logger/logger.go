package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Level string

const (
	InfoLevel  Level = "INFO"
	ErrorLevel Level = "ERROR"
	DebugLevel Level = "DEBUG"
	WarnLevel  Level = "WARN"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

type Fields map[string]interface{}

type Logger struct {
	mu           sync.Mutex
	infoLogger   *log.Logger
	errorLogger  *log.Logger
	debugLogger  *log.Logger
	warnLogger   *log.Logger
	contextField Fields
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:   log.New(os.Stdout, "INFO: ", 0),
		errorLogger:  log.New(os.Stderr, "ERROR: ", 0),
		debugLogger:  log.New(os.Stdout, "DEBUG: ", 0),
		warnLogger:   log.New(os.Stdout, "WARN: ", 0),
		contextField: make(Fields),
	}
}

// WithContext creates a new logger with context
func WithContext(ctx context.Context) *Logger {
	logger := NewLogger()

	// Extract request ID or trace ID if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		logger.contextField["request_id"] = requestID
	}

	return logger
}

// WithFields adds additional context fields to the logger
func (l *Logger) WithFields(fields Fields) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a copy of existing fields and merge new fields
	newFields := make(Fields)
	for k, v := range l.contextField {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	l.contextField = newFields
	return l
}

// WithError adds an error to the context fields
func (l *Logger) WithError(err error) *Logger {
	if err != nil {
		return l.WithFields(Fields{"error": err.Error()})
	}
	return l
}

func (l *Logger) Log(level Level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	var colorCode string
	switch level {
	case InfoLevel:
		colorCode = colorGreen
	case ErrorLevel:
		colorCode = colorRed
	case DebugLevel:
		colorCode = colorBlue
	case WarnLevel:
		colorCode = colorYellow
	}
	contextStr := ""
	if len(l.contextField) > 0 {
		contextStr = " " + formatContextFields(l.contextField)
	}

	coloredLog := fmt.Sprintf("%s[%s] [%s]%s %s%s\n",
		colorCode,
		timestamp,
		level,
		contextStr,
		msg,
		colorReset)

	switch level {
	case InfoLevel:
		fmt.Print(coloredLog)
	case ErrorLevel:
		fmt.Fprint(os.Stderr, coloredLog)
	case DebugLevel:
		fmt.Print(coloredLog)
	case WarnLevel:
		fmt.Print(coloredLog)
	}
}

// Convenience methods for different log levels
func (l *Logger) Info(format string, v ...interface{}) {
	l.Log(InfoLevel, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Log(ErrorLevel, format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.Log(DebugLevel, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.Log(WarnLevel, format, v...)
}

func formatContextFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	var fieldsStr string
	for k, v := range fields {
		fieldsStr += fmt.Sprintf(" %s=%v", k, v)
	}
	return fieldsStr
}
