package logger

import (
	"fmt"
	"log"
)

const (
	Reset      = "\033[0m"
	RedBright  = "\033[91m"
	CyanBright = "\033[96m"
)

func logMessage(color, label, symbol, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	log.Printf("%s%s%s %s %s", color, symbol, label, message, Reset)
}

func Info(format string, args ...any) {
	logMessage(CyanBright, "INFO", "ℹ️", format, args...)
}

func Error(format string, args ...any) {
	logMessage(RedBright, "ERROR", "❌", format, args...)
}

func Fatal(err error, msg string) {
	if err != nil {
		log.Fatalf(
			"%s☠ FATAL [%s: %s]%s",
			RedBright,
			msg,
			err,
			Reset,
		)
	}
}

func Start() {
	log.Printf(`%s
#################################################################
                        Motif API Server
#################################################################
%s`, CyanBright, Reset)
}
