package windows

import "time"

// Win32_Service from https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-service
type Win32_Service struct {
	AcceptPause             bool
	AcceptStop              bool
	Caption                 string
	CheckPoint              uint32
	CreationClassName       string
	DelayedAutoStart        bool
	Description             string
	DesktopInteract         bool
	DisplayName             string
	ErrorControl            string
	ExitCode                uint32
	InstallDate             time.Duration
	Name                    string
	PathName                string
	ProcessId               uint32
	ServiceSpecificExitCode uint32
	ServiceType             string
	Started                 bool
	StartMode               string
	StartName               string
	State                   string
	Status                  string
	SystemCreationClassName string
	SystemName              string
	TagId                   uint32
	WaitHint                uint32
}
