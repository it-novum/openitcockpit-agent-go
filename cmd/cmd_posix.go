// +build linux darwin

package cmd

import "os"

func PlatformMain() {
	if err := New().Execute(); err != nil {
		os.Exit(1)
	}
}
