package checks

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/evlog"
)

type resultEvent struct {
	Message     string
	Channel     string
	Level       string
	LevelRaw    uint8
	RecordID    uint64
	TimeCreated time.Time
	Provider    string
	Task        string
	Keywords    []string
}

type CheckWindowsEventLog struct {
	eventLogs  map[string]*evlog.EventLog
	cache      time.Duration
	eventCache map[string][]*resultEvent
}

// Name will be used in the response as check name
func (c *CheckWindowsEventLog) Name() string {
	return "windows_eventlog"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWindowsEventLog) Run(ctx context.Context) (interface{}, error) {
	for channel, eventLog := range c.eventLogs {
		events, err := eventLog.Query()
		if err != nil {
			log.Errorln("Event Log: could not query event log: ", err)
			return nil, err
		}
		eventCache := c.eventCache[channel]

		for _, event := range events {
			eventCache = append(eventCache, &resultEvent{
				Message:     event.Message,
				Channel:     event.Channel,
				Level:       event.Level,
				LevelRaw:    event.LevelRaw,
				RecordID:    event.RecordID,
				TimeCreated: event.TimeCreated.SystemTime,
				Provider:    event.Provider.Name,
				Task:        event.Task,
				Keywords:    event.Keywords,
			})
		}

		log.Debugln("Event Log: cache before cleanup: ", len(eventCache))
		if len(eventCache) > 0 {
			firstIndex := 0
			keepTime := time.Now().Add(-c.cache)
			for i, event := range eventCache {
				if event.TimeCreated.After(keepTime) {
					firstIndex = i
					break
				}
			}
			eventCache = eventCache[firstIndex:]
		}
		log.Debugln("Event Log: cache after cleanup: ", len(eventCache))

		c.eventCache[channel] = eventCache
	}

	return c.eventCache, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckWindowsEventLog) Configure(cfg *config.Configuration) (bool, error) {
	if cfg.WindowsEventLog && len(cfg.WindowsEventLogTypes) > 0 {
		avail, err := evlog.Available()
		if err != nil {
			return false, fmt.Errorf("windows event log availability check: %s", err)
		}
		if !avail {
			return false, fmt.Errorf("windows event log is not available")
		}

		var cacheSec uint64 = 3600
		if cfg.WindowsEventLogCache > 0 {
			cacheSec = uint64(cfg.WindowsEventLogCache)
		}
		c.cache = time.Second * time.Duration(cacheSec)
		c.eventCache = make(map[string][]*resultEvent)
		c.eventLogs = make(map[string]*evlog.EventLog)

		for _, channel := range cfg.WindowsEventLogTypes {
			c.eventCache[channel] = make([]*resultEvent, 0)
			c.eventLogs[channel] = &evlog.EventLog{
				LogChannel: channel,
				TimeDiff:   cacheSec,
			}
		}

		return true, nil
	}
	return false, nil
}
