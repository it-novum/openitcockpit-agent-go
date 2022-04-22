//go:build linux || darwin
// +build linux darwin

package basiclog

import (
	"log/syslog"
)

// This very basic log gets used to log any config parser errors
// to the syslog or Windows Event Log on systems that do not have a syslog.
// This is because if the agent can not parse it's config.ini it has no logfile path.

type BasicLogger struct {
	handler *syslog.Writer
}

func New() (*BasicLogger, error) {

	logHandle, err := syslog.New(syslog.LOG_ERR, "openITCOCKPITAgent")
	if err != nil {
		return nil, err
	}

	return &BasicLogger{
		handler: logHandle,
	}, nil
}

func (l *BasicLogger) LogError(msg string) error {
	return l.handler.Err(msg)
}
