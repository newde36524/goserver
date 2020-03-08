package goserver

import (
	"fmt"
)

var (
	logTemp      = "[goserver]:%s\n"
	errlogTemp   = "[Error]" + logTemp + "%s"
	infologTemp  = "[Info]" + logTemp + "%s"
	debuglogTemp = "[Debug]" + logTemp + "%s"
)

//logInfo .
func logInfo(msg string) {
	fmt.Printf(infologTemp, logTemp, msg)
}

//logDebug .
func logDebug(msg string) {
	fmt.Printf(debuglogTemp, logTemp, msg)
}

//logError .
func logError(msg string) {
	fmt.Printf(errlogTemp, logTemp, msg)
}

func panicError(msg string) {
	panic(fmt.Sprintf(errlogTemp, msg))
}
