package logger

import (
	"log"
	"os"
)

// Logger is a simple logger interface
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	debugEnabled bool
}

// New creates a new logger instance
func New() *Logger {
	// Check if debug logs are enabled (default: true)
	debugEnabled := os.Getenv("DEBUG_LOGS_ENABLED") != "false"

	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugEnabled: debugEnabled,
	}
}

// Info logs an info message (controlled by DEBUG_LOGS_ENABLED)
func (l *Logger) Info(v ...interface{}) {
	if l.debugEnabled {
		l.infoLogger.Println(v...)
	}
}

// Infof logs a formatted info message (controlled by DEBUG_LOGS_ENABLED)
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.debugEnabled {
		l.infoLogger.Printf(format, v...)
	}
}

// Error logs an error message (always enabled)
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Errorf logs a formatted error message (always enabled)
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs a debug message (controlled by DEBUG_LOGS_ENABLED)
func (l *Logger) Debug(v ...interface{}) {
	if l.debugEnabled {
		l.debugLogger.Println(v...)
	}
}

// Debugf logs a formatted debug message (controlled by DEBUG_LOGS_ENABLED)
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.debugEnabled {
		l.debugLogger.Printf(format, v...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLogger.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}
