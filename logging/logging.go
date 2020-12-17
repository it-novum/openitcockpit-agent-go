package logging

import (
	"context"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

type LogHandler struct {
	LogPath string
	Rotate  uint
	Verbose bool

	logFile *os.File
	reload  chan struct{}
}

func init() {
	log.SetLevel(log.InfoLevel)
}

// Reload triggers an async reload of the log file (reopen)
func (h *LogHandler) Reload() {
	h.reload <- struct{}{}
}

func (h *LogHandler) doReload() {
	if h.Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if h.LogPath != "" {
		fl, err := os.OpenFile(h.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Fatal("Could not open/create log file: ", err)
		}
		log.SetOutput(io.MultiWriter(os.Stderr, fl))
		if h.logFile != nil {
			h.logFile.Close()
		}
		h.logFile = fl
	}
}

func (h *LogHandler) doRotate() {
	if h.Rotate > 0 && h.logFile != nil {
		if h.Rotate > 1 {
			for i := h.Rotate - 1; i > 1; i-- {
				//TODO do rotate
			}
		}
	}
}

// Run the log handling (should be run in a go routine)
func (h *LogHandler) Run(ctx context.Context) {
	defer func() {
		close(h.reload)
		h.Close()
	}()

	// TODO call doRotate on noon

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.reload:
			h.doReload()
		}
	}
}

// Close all files
func (h *LogHandler) Close() {
	if h.logFile != nil {
		log.SetOutput(os.Stderr)
		h.logFile.Close()
		h.logFile = nil
	}
}
