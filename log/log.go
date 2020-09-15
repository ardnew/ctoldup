package log

import "fmt"

const (
	Info = iota
	Error
)

func Msg(level int, prompt string, format string, args ...interface{}) {
	var prefix string
	switch level {
	case Info:
		prefix = "[ ]"
	case Error:
		prefix = "[!]"
	}
	fmt.Printf(fmt.Sprintf("%s %s: %s\n", prefix, prompt, format), args...)
}
