package utils

import (
	"bufio"
	"os"
)

func ReadUserPrompt() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}
