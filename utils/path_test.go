package utils

import (
	"runtime"
	"testing"
)

func TestFileExists(t *testing.T) {
	path := "/etc/hosts"
	npath := "/sdfsdfdf"
	if runtime.GOOS == "windows" {
		path = `C:\Windows\explorer.exe`
		npath = `C:\sdfsfsdf`
	}
	if !FileExists(path) && FileNotExists(path) {
		t.Error("FileExists doesn't find existing file")
	}
	if FileExists(npath) && !FileNotExists(path) {
		t.Error("FileExists finds non existing file")
	}
}
