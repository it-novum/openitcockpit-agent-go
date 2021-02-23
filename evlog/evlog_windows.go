package evlog

import (
	"bytes"

	evsys "github.com/elastic/beats/v7/winlogbeat/sys"
	"github.com/elastic/beats/v7/winlogbeat/sys/wineventlog"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

type EventLog struct {
	LogChannel []string
	Bookmark   *evsys.Event
}

func (e *EventLog) prepare() (string, wineventlog.EvtHandle, wineventlog.EvtSubscribeFlag, error) {
	var (
		err      error
		bookmark wineventlog.EvtHandle
		flags    = wineventlog.EvtSubscribeStartAtOldestRecord
	)
	log.Debugln("evlog: prepare query")

	timeDiff := 3600
	if e.Bookmark != nil {
		timeDiff = 0
		bookmark, err = wineventlog.CreateBookmarkFromRecordID(e.Bookmark.Channel, e.Bookmark.RecordID)
		if err != nil {
			log.Errorln(errors.Wrap(err, "evlog: could not create bookmark"))
			bookmark = 0
		} else {
			flags = wineventlog.EvtSubscribeStartAfterBookmark
		}
	}

	query, err := queryStringFromChannels(e.LogChannel, timeDiff)
	if err != nil {
		return "", 0, 0, errors.Wrap(err, "evlog: could not create query xml")
	}
	return query, bookmark, flags, nil
}

func (e *EventLog) renderEvent(eventHandle wineventlog.EvtHandle) (*evsys.Event, error) {
	var (
		rendered = false
		bufSize  = 1024
		output   = bytes.Buffer{}
	)

	for !rendered {
		renderBuf := make([]byte, bufSize)
		err := wineventlog.RenderEventXML(eventHandle, renderBuf, &output)
		if err == nil {
			rendered = true
		} else {
			_, bufErr := err.(evsys.InsufficientBufferError)
			if bufErr {
				bufSize *= 2
				output = bytes.Buffer{}
			} else {
				return nil, errors.Wrap(err, "evlog: could not render event")
			}
		}
	}

	event, err := evsys.UnmarshalEventXML(output.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "evlog: could not unmarshal event xml")
	}

	return &event, nil
}

func (e *EventLog) Query() ([]evsys.Event, error) {
	query, bookmark, flags, err := e.prepare()
	if err != nil {
		return nil, err
	}

	log.Debugln("evlog: create windows event handle")
	signalHandle, err := windows.CreateEvent(nil, 1, 1, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create windows event handle")
	}
	defer windows.CloseHandle(signalHandle)

	log.Debugln("evlog: subscribe to windows event log with query: ", query)
	subscription, err := wineventlog.Subscribe(0, signalHandle, "", query, bookmark, flags)
	if err != nil {
		return nil, errors.Wrap(err, "evlog: could not subscribe to event log")
	}
	defer subscription.Close()

	log.Debugln("evlog: fetch event log handles")
	iter, err := wineventlog.NewEventIterator(wineventlog.WithSubscription(subscription))
	if err != nil {
		return nil, errors.Wrap(err, "evlog: could not create event iterator from subscription")
	}
	defer iter.Close()

	var (
		events []evsys.Event
	)

	for {
		eventHandle, more := iter.Next()
		if !more {
			break
		}

		event, err := e.renderEvent(eventHandle)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	if err := iter.Err(); err != nil {
		return nil, errors.Wrap(err, "evlog: could not fetch events from subscription")
	}
	log.Debugln("evlog: fetched and parsed all events: ", len(events))

	if len(events) > 0 {
		e.Bookmark = &(events[len(events)-1])
	}

	return events, nil
}

func Available() (bool, error) {
	return wineventlog.IsAvailable()
}
