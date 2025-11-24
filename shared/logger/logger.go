package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

// Logger provides structured logging for services
type Logger struct {
	serviceName string
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger creates a new logger instance for a service
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
		infoLogger:  log.New(os.Stdout, fmt.Sprintf("INFO: [%s] ", serviceName), log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, fmt.Sprintf("ERROR: [%s] ", serviceName), log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.infoLogger.Println(message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.errorLogger.Println(message)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string) {
	l.errorLogger.Fatal(message)
}

// Deprecated: Use NewLogger instead
func Init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Deprecated: Use Logger.Info instead
func Info(v ...interface{}) {
	InfoLogger.Println(v...)
}

// Deprecated: Use Logger.Error instead
func Error(v ...interface{}) {
	ErrorLogger.Println(v...)
}
