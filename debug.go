package gobot

import "fmt"

func Debuglnf(isDebug bool, format string, a ...any) {
	if isDebug {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf("<Debug>: %s\n", msg)
	}
}
