package protocol

import "fmt"

func Error(msg string) string {
	return "-ERR " + msg + "\r\n"
}

func OK() string {
	return "+OK" + "\r\n"
}

func BulkString(s string) string {
	if s == "" {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func NullBulk() string {
	return "$-1\r\n"
}

func WriteSimpleString(msg string) string {
	return fmt.Sprintf("+%s\r\n", msg)
}
