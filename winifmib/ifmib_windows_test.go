package winifmib

import (
	"encoding/hex"
	"encoding/json"
	"golang.org/x/sys/windows"
	"reflect"
	"testing"
	"unsafe"
)

func TestGetIfEntry2Ex(t *testing.T) {
	row, err := GetIfEntry2Ex(26, true)
	if err != nil {
		t.Fatal(err)
	}
	js, err := json.MarshalIndent(row, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(js))
}

func TestSizeof(t *testing.T) {
	/* # tested in visual studio 2019 x86 and x64
	#include <WinSock2.h>
	#include <netioapi.h>
	#include <windows.h>
	#include <stdio.h>
	int main(void) {
	    MIB_IF_ROW2 t;
	    printf("MIB_IF_ROW2: %d", sizeof(t));
	    return 0;
	}
	*/
	expected := uintptr(1352)

	if MibIfRow2Size != expected {
		t.Fatal("unexpected size: ", MibIfRow2Size)
	}
}

func hexDump(row *MibIfRow2) string {
	size := unsafe.Sizeof(*row)
	var sl []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&sl))
	sliceHeader.Cap = int(size)
	sliceHeader.Len = int(size)
	sliceHeader.Data = uintptr(unsafe.Pointer(row))
	return hex.Dump(sl)
}

func TestGetIfTable2Ex(t *testing.T) {
	table, err := GetIfTable2Ex(true)
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()
	for i := Ulong(0); i < table.NumEntries; i++ {
		row := table.Get(i)
		t.Log("Index:", row.InterfaceIndex)
		t.Log("Alias:", windows.UTF16ToString(row.Alias[:]))
		t.Log("Desc:", windows.UTF16ToString(row.Description[:]))
		t.Log("Type:", row.Type)
		t.Log("Mtu:", row.Mtu)
		//t.Log(hexDump(row))
		t.Log("")
	}
	t.Fail()
}

func TestEthernet(t *testing.T) {
	table, err := GetIfTable2Ex(true)
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()

	rows := table.Slice()
	for _, row := range rows {
		name := windows.UTF16ToString(row.Alias[:])
		if name == "Ethernet" {
			js, err := json.MarshalIndent(row, "", "    ")
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(js))
		}
	}
}

func TestNetworkInterfaceStatistics(t *testing.T) {
	stats, err := NetworkInterfaceStatistics()
	if err != nil {
		t.Fatal(err)
	}

	for _, stat := range stats {
		t.Log(stat.Name, ": RX: ", stat.BytesReceived, " TX: ", stat.BytesSent)
	}
}
