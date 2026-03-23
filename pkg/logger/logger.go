package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	stdoutLogger *log.Logger
	stderrLogger *log.Logger
)

func init() {
	Init()
}

func Init() {
	stdoutLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	stderrLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func Log(level, category, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	prefix := fmt.Sprintf("[%s] [%s:] ", category, level)

	switch level {
	case "ERROR":
		stderrLogger.Printf("%s%s", prefix, msg)
	default:
		stdoutLogger.Printf("%s%s", prefix, msg)
	}
}
