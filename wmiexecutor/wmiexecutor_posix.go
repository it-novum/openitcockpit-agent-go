// +build linux darwin

package wmiexecutor

import (
	"fmt"
	"os"
)

func PlatformMain() {
	fmt.Println("This programm can only run on Windows Systems.")
	os.Exit(0)
}
