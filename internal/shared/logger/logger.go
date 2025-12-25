package logger

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

var instance *Logger

// Init initializes the logger
func Init() {
	instance = &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs info messages
func Info(format string, v ...interface{}) {
	if instance == nil {
		Init()
	}
	instance.infoLogger.Printf(format, v...)
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	if instance == nil {
		Init()
	}
	instance.errorLogger.Printf(format, v...)
}

// Debug logs debug messages
func Debug(format string, v ...interface{}) {
	if instance == nil {
		Init()
	}
	instance.debugLogger.Printf(format, v...)
}
