package pushclient

import "testing"

func TestAddressForIPPort(t *testing.T) {
	testMap := map[string]string{
		"192.168.1.1":   "192.168.1.1:34234",
		"[2001:db8::1]": "[2001:db8::1]:49151",
	}

	for expected, value := range testMap {
		result := addressForIPPort(value)
		if expected != result {
			t.Error("Value: ", value, " Result: ", result, " Expected: ", expected)
		}
	}
}

func TestFetchSystemInformation(t *testing.T) {
	hostname, ip := fetchSystemInformation()
	if hostname == "" {
		t.Error("Expected hostname")
	}
	if ip == "" {
		t.Error("Expected ip address")
	}
}
