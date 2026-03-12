package logger

import (
	"log"
	"os"
)

var (
	infoLogger     *log.Logger
	errorLogger    *log.Logger
	registryLogger *log.Logger
	httpLogger     *log.Logger
	presenceLogger *log.Logger
	netLogger      *log.Logger
	appLogger      *log.Logger
)

func init() {
	Init()
}

func Init() {
	infoLogger = log.New(os.Stdout, "[INFO:] ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR:] ", log.Ldate|log.Ltime)

	registryLogger = log.New(os.Stdout, "[REGISTRY:] ", log.Ldate|log.Ltime)
	httpLogger = log.New(os.Stdout, "[HTTP:] ", log.Ldate|log.Ltime)
	appLogger = log.New(os.Stdout, "[APP:] ", log.Ldate|log.Ltime)
	netLogger = log.New(os.Stdout, "[NET:] ", log.Ldate|log.Ltime)
	presenceLogger = log.New(os.Stdout, "[PRESENCE:] ", log.Ldate|log.Ltime)

}

func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}
func Registry(format string, v ...interface{}) {
	registryLogger.Printf(format, v...)
}
func HTTP(format string, v ...interface{}) {
	httpLogger.Printf(format, v...)
}
func App(format string, v ...interface{}) {
	appLogger.Printf(format, v...)
}
func Net(format string, v ...interface{}) {
	netLogger.Printf(format, v...)
}
func Presence(format string, v ...interface{}) {
	presenceLogger.Printf(format, v...)
}
