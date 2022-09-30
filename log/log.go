package log

import "time"

func Timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Info(msg string) {
	println("[" + Timestamp() + "] [INFO] " + msg)
}

func Error(msg string) {
	println("[" + Timestamp() + "] [ERROR] " + msg)
}
