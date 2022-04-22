package basiclog

import (
	"golang.org/x/sys/windows/svc/eventlog"
)

// This very basic log gets used to log any config parser errors
// to the syslog or Windows Event Log on systems that do not have a syslog.
// This is because if the agent can not parse it's config.ini it has no logfile path.

type BasicLogger struct {
	handler *eventlog.Log
}

func New() (*BasicLogger, error) {

	logHandle, err := eventlog.Open("openITCOCKPITAgent")
	if err != nil {
		return nil, err
	}

	return &BasicLogger{
		handler: logHandle,
	}, nil
}

func (l *BasicLogger) LogError(msg string) error {
	return l.handler.Error(1, msg)
}
