package logging

import (
	"fmt"
	"time"
	"os"
)

const reset = "\033[0m"
const red = "\033[31m"
const green = "\033[32m"
const yellow = "\033[33m"
const blue = "\033[34m"
const magenta = "\033[35m"
const cyan = "\033[36m"
const gray = "\033[37m"
const white = "\033[97m"

// 3 = ERROR
// 2 = WARN
// 1 = INFO
// 0 = TRACE
var logLevel = 3;

func SetLogLevel(level int) {
	logLevel = level
}

func timeFormat(t time.Time) (r string) {
	r = fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return r
}


func Fatal(message ...any) {
	fmt.Printf("%s%s | FATAL:\t", red, timeFormat(time.Now()))
	for _, msg := range message {
		fmt.Print(msg)
	}

	fmt.Printf("%s\n", reset)
	os.Exit(1)
}

// error out
func Error(message ...any) {
	if logLevel == 3 {
		fmt.Printf("%s%s | ERR:\t", red, timeFormat(time.Now()))
		for _, msg := range message {
			fmt.Print(msg)
		}

		fmt.Printf("%s\n", reset)
	}
	
}

func Warn(message ...any) {
	if logLevel >= 2 {
		fmt.Printf("%s%s | WARN:\t", yellow, timeFormat(time.Now()))
		for _, msg := range message {
			fmt.Print(msg)
		}

		fmt.Printf("%s\n", reset)
	}
}

func Info(message ...any) {
	if logLevel >= 1 {

	}
	fmt.Printf("%s%s | INFO:\t", blue, timeFormat(time.Now()))
	for _, msg := range message {
		fmt.Print(msg)
	}

	fmt.Printf("%s\n", reset)
}

func Trace(message ...any) {
	if logLevel == 0 {
		fmt.Printf("%s%s | TRACE:\t", blue, timeFormat(time.Now()))
		for _, msg := range message {
			fmt.Print(msg)
		}

		fmt.Printf("%s\n", reset)
	}
	
}
