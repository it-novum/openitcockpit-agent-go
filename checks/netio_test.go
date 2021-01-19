package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestChecksCheckNetIO(t *testing.T) {

	check := &CheckNetIo{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok := cr.(map[string]*resultNetIo)
	if !ok {
		t.Fatal("False type")
	}

	var oldPacketsSent []uint64
	for _, result := range results {
		fmt.Printf("Nic [Check 1]: %s\n", result.Name)
		fmt.Printf("Packets sent: %v\n", result.PacketsSent)
		fmt.Printf("Packets Received: %v\n", result.PacketsReceived)
		oldPacketsSent = append(oldPacketsSent, result.PacketsSent)
	}

	time.Sleep(10 * time.Second)

	cr, err = check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok = cr.(map[string]*resultNetIo)
	if !ok {
		t.Fatal("False type")
	}

	var newPacketsSent []uint64
	var js []byte
	for _, result := range results {
		fmt.Printf("Nic [Check 2]: %s\n", result.Name)
		fmt.Printf("Packets sent: %v\n", result.PacketsSent)
		fmt.Printf("Packets Received: %v\n", result.PacketsReceived)

		//js, _ = json.Marshal(result)
		//fmt.Println(string(js))

		newPacketsSent = append(newPacketsSent, result.PacketsSent)
	}

	wasTrafficOnOneInterface := false
	for i, value := range oldPacketsSent {
		if newPacketsSent[i] >= value {
			wasTrafficOnOneInterface = true
		}

	}
	if !wasTrafficOnOneInterface {
		t.Fatal("No packets send on all interfaces - that's suspicious")
	}

	js, _ = json.Marshal(results)
	fmt.Println(string(js))

}
