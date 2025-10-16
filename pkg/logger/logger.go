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
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLogger.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}
