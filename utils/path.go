package utils

import "os"

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func FileNotExists(path string) bool {
	return !FileExists(path)
}
