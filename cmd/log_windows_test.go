package cmd

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestPlatformLogFile(t *testing.T) {
	testLogPath = ""
	registryKey = registry.CURRENT_USER
	registryPath = `SOFTWARE\OITCTEST`

	key, _, err := registry.CreateKey(registryKey, registryPath, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal("Could not create test registry key: ", err)
	}
	closed := false
	defer func() {
		if !closed {
			key.Close()
		}
		if err := registry.DeleteKey(registryKey, registryPath); err != nil {
			t.Fatal("could not delete test registry key: ", err)
		}
	}()

	if err := key.SetStringValue(registryName, "test"); err != nil {
		t.Fatal("could not set registry test value: ", err)
	}
	if err := key.Close(); err != nil {
		closed = true
		t.Fatal("could not close registry key: ", err)
	}
	closed = true

	test := platformLogFile()
	if test != "test/agent.log" {
		t.Error("platformLogFile did not return correct registry value: ", test)
	}
}
