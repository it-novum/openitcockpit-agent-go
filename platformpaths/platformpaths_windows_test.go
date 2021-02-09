package platformpaths

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestPlatformRegistryRead(t *testing.T) {
	wpp := getPlatformPath().(*windowsPlatformPath)

	wpp.registryKey = registry.CURRENT_USER
	wpp.registryPath = `SOFTWARE\OITCTEST`

	key, _, err := registry.CreateKey(wpp.registryKey, wpp.registryPath, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal("Could not create test registry key: ", err)
	}
	closed := false
	defer func() {
		if !closed {
			key.Close()
		}
		if err := registry.DeleteKey(wpp.registryKey, wpp.registryPath); err != nil {
			t.Fatal("could not delete test registry key: ", err)
		}
	}()

	if err := key.SetStringValue(wpp.registryName, "test"); err != nil {
		t.Fatal("could not set registry test value: ", err)
	}
	if err := key.SetStringValue("Verbose", "1"); err != nil {
		t.Fatal("could not set registry Verbose value: ", err)
	}
	// test if invalid type crashes program, this should just be logged and ignored later
	if err := key.SetDWordValue("Debug", 1); err != nil {
		t.Fatal("could not set registry Debug value: ", err)
	}
	if err := key.Close(); err != nil {
		closed = true
		t.Fatal("could not close registry key: ", err)
	}
	closed = true

	if err := wpp.Init(); err != nil {
		t.Error(err)
	}
	test := wpp.LogPath()
	if test != "test\\agent.log" {
		t.Error("PlatformPath did not return correct registry value: ", test)
	}
	testVerbose, ok := wpp.AdditionalData()["Verbose"]
	if !ok {
		t.Error("PlatformPath could not fetch Verbose from registry")
	} else if testVerbose != "1" {
		t.Error("PlatformPath did not return correct registry value for Verbose: ", testVerbose)
	}
	testDebug, ok := wpp.AdditionalData()["Debug"]
	if ok {
		t.Error("PlatformPath did return registry value for Debug but should ignore it: ", testDebug)
	}
}
