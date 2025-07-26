package logger

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func Init(level string) {
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	if infoLogger != nil {
		infoLogger.Println(msg)
	}
}

func Error(msg string, err error) {
	if errorLogger != nil {
		if err != nil {
			errorLogger.Printf("%s: %v", msg, err)
		} else {
			errorLogger.Println(msg)
		}
	}
}