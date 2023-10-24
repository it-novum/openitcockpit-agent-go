package utils

import (
	"io/fs"

	"github.com/hectane/go-acl"
	"golang.org/x/sys/windows"
)

// This version of Chmod can be used to set filepermissions for the current user and read and write access for the SYSTEM user
// This func is only relevant for Windows
func Chmod(name string, mode fs.FileMode) error {
	if err := acl.Chmod(name, mode); err != nil {
		return err
	}

	// Set read and write permissions for the SYSTEM user
	if err := acl.Apply(
		name,
		false,
		false,
		acl.GrantName(windows.GENERIC_READ, "SYSTEM"),
		acl.GrantName(windows.GENERIC_WRITE, "SYSTEM"),
	); err != nil {
		return err
	}

	return nil
}
