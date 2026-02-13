package logger

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	Init()
}

func Init() {
	infoLogger = log.New(os.Stdout, "[INFO:] ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stdout, "[ERROR:] ", log.Ldate|log.Ltime)
}

func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}
