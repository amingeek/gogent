package utils

import (
	"bufio"
	"log"
	"os"
)

func ReadUserPrompt() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}

func Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
