package goserver

import (
	"log"
	"os"
	"runtime"
)

var (
	logger       = log.New(os.Stdout, "[goserver]", 0)
	infologTemp  = "[Info]%s:%d: %s\n"
	debuglogTemp = "[Debug]%s:%d: %s\n"
	errlogTemp   = "[Error]%s:%d: %s\n"
)

func getCallerDesc() (file string, line int) {
	_, file, line, _ = runtime.Caller(2)
	return
}

//logInfo .
func logInfo(msg string) {
	file, line := getCallerDesc()
	logger.Printf(infologTemp, file, line, msg)
}

//logDebug .
func logDebug(msg string) {
	file, line := getCallerDesc()
	logger.Printf(debuglogTemp, file, line, msg)
}

//logError .
func logError(msg string) {
	file, line := getCallerDesc()
	logger.Printf(errlogTemp, file, line, msg)
}

func panicError(msg string) {
	file, line := getCallerDesc()
	logger.Panicf(errlogTemp, file, line, msg)
}
