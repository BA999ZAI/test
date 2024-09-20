package utils

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func LogMessage(format string, v ...interface{}) {
	logger.Printf(format, v...)
}