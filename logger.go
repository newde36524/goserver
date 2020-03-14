package goserver

import (
	"fmt"
	"runtime"
)

var (
	infologTemp  = "[goserver][Info]%s:%d: %s\n"
	debuglogTemp = "[goserver][Debug]%s:%d: %s\n"
	errlogTemp   = "[goserver][Error]%s:%d: %s\n"
)

func getCallerDesc() (file string, line int) {
	_, file, line, _ = runtime.Caller(2)
	return
}

//logInfo .
func logInfo(msg string) {
	file, line := getCallerDesc()
	fmt.Printf(infologTemp, file, line, msg)
}

//logDebug .
func logDebug(msg string) {
	file, line := getCallerDesc()
	fmt.Printf(debuglogTemp, file, line, msg)
}

//logError .
func logError(msg string) {
	file, line := getCallerDesc()
	fmt.Printf(errlogTemp, file, line, msg)
}

func panicError(msg string) {
	file, line := getCallerDesc()
	panic(fmt.Sprintf(errlogTemp, file, line, msg))
}
