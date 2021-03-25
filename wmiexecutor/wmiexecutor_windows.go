package wmiexecutor

import (
	"fmt"
	"os"
)

func PlatformMain() {

	if err := New().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
