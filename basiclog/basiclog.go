package basiclog

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// This very basic log gets used to log any config parser errors
// to the syslog or Windows Event Log on systems that do not have a syslog.
// This is because if the agent can not parse it's config.ini it has no logfile path.

func (l *BasicLogger) Errorln(args ...interface{}) error {
	msg := fmt.Sprint(args...)

	t := time.Now()
	source := l.source()
	template := "time=\"%v\" level=error msg=\"%v\" source=\"%v\"\n"

	msg = fmt.Sprintf(template, t.Format(time.RFC3339), msg, source)

	// Print msg
	fmt.Print(msg)

	if l.handler == nil {
		return fmt.Errorf("No log handler initialized! Message will only be printed to stdout.")
	}

	// Save msg to logfile via given handler
	return l.LogError(msg)
}

func (l *BasicLogger) source() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}

	return fmt.Sprintf("%s:%d", file, line)
}
