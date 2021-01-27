// +build windows

package winapi

import (
	"fmt"
	"time"
)

const (
	SESS_INTERACTIVE_LOGON        = 2
	SESS_REMOTE_INTERACTIVE_LOGON = 10
	SESS_CACHED_INTERACTIVE_LOGON = 11
)

type SessionDetails struct {
	Username      string    `json:"username"`
	Domain        string    `json:"domain"`
	LocalUser     bool      `json:"isLocal"`
	LocalAdmin    bool      `json:"isAdmin"`
	LogonType     uint32    `json:"logonType"`
	LogonTime     time.Time `json:"logonTime"`
	DnsDomainName string    `json:"dnsDomainName"`
}

func (s *SessionDetails) FullUser() string {
	return fmt.Sprintf("%s\\%s", s.Domain, s.Username)
}

func (s *SessionDetails) GetLogonType() string {
	switch s.LogonType {
	case SESS_INTERACTIVE_LOGON:
		return "INTERACTIVE_LOGON"
	case SESS_REMOTE_INTERACTIVE_LOGON:
		return "REMOTE_INTERACTIVE_LOGON"
	case SESS_CACHED_INTERACTIVE_LOGON:
		return "CACHED_INTERACTIVE_LOGON"
	default:
		return "UNKNOWN"
	}
}

type Process struct {
	Pid        int    `json:"pid"`
	Ppid       int    `json:"parentpid"`
	Executable string `json:"exeName"`
	Fullpath   string `json:"fullPath"`
	Username   string `json:"username"`
}

type LocalUser struct {
	Username             string        `json:"username"`
	FullName             string        `json:"fullName"`
	IsEnabled            bool          `json:"isEnabled"`
	IsLocked             bool          `json:"isLocked"`
	IsAdmin              bool          `json:"isAdmin"`
	PasswordNeverExpires bool          `json:"passwordNeverExpires"`
	NoChangePassword     bool          `json:"noChangePassword"`
	PasswordAge          time.Duration `json:"passwordAge"`
	LastLogon            time.Time     `json:"lastLogon"`
	BadPasswordCount     uint32        `json:"badPasswordCount"`
	NumberOfLogons       uint32        `json:"numberOfLogons"`
}
