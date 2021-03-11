// Code generated by 'go generate'; DO NOT EDIT.

package winifmib

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modIphlpapi = windows.NewLazySystemDLL("Iphlpapi.dll")

	procFreeMibTable  = modIphlpapi.NewProc("FreeMibTable")
	procGetIfEntry2Ex = modIphlpapi.NewProc("GetIfEntry2Ex")
	procGetIfTable2Ex = modIphlpapi.NewProc("GetIfTable2Ex")
)

func freeMibTable(memory uintptr) {
	syscall.Syscall(procFreeMibTable.Addr(), 1, uintptr(memory), 0, 0)
	return
}

func getIfEntry2Ex(level MibIfEntryLevel, row *MibIfRow2) (err error) {
	r1, _, e1 := syscall.Syscall(procGetIfEntry2Ex.Addr(), 2, uintptr(level), uintptr(unsafe.Pointer(row)), 0)
	if r1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func getIfTable2Ex(level MibIfEntryLevel, table *PMibIfTable) (err error) {
	r1, _, e1 := syscall.Syscall(procGetIfTable2Ex.Addr(), 2, uintptr(level), uintptr(unsafe.Pointer(table)), 0)
	if r1 != 0 {
		err = errnoErr(e1)
	}
	return
}
