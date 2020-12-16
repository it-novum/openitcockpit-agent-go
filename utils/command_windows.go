package utils

import (
	"os"
	"strings"
)

func handleCommandError(arg0 string, err error) int {
	if strings.HasSuffix(err.Error(), "file does not exist") {
		if _, err := os.Stat(arg0); os.IsNotExist(err) {
			return NotFound
		}
		return NotExecutable
	}
	return Unknown
}
