// +build !windows

package utils

import (
	"os"
)

func handleCommandError(arg0 string, err error) int {
	if os.IsNotExist(err) { // does not work with windows
		return NotFound
	}

	if os.IsPermission(err) { // does not work with windows
		return NotExecutable
	}

	return Unknown
}
