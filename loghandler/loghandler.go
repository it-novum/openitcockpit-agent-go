package loghandler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// LogHandler cares about files and rotation
type LogHandler struct {
	// path to log file
	LogPath string
	// value < 1 will disable logrotate
	LogRotate int
	// usually os.Stderr
	DefaultWriter io.Writer
	Verbose       bool
	Debug         bool

	wg       sync.WaitGroup
	logFile  *os.File
	shutdown chan struct{}
}

func (h *LogHandler) openLogFile() {
	log.Infoln("LogHandler: create/open log file")

	fl, err := os.OpenFile(h.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("LogHandler: could not open/create log file: ", err)
	}
	log.SetOutput(io.MultiWriter(h.DefaultWriter, fl))
	log.Infoln("LogHandler: log file opened successfully: ", h.LogPath)

	h.logFile = fl
}

func (h *LogHandler) closeLogFile() {
	log.SetOutput(h.DefaultWriter)

	if h.logFile != nil {
		if err := h.logFile.Close(); err != nil {
			log.Errorln("LogHandler: could not close log file: ", err)
		}
		h.logFile = nil
	}
}

func (h *LogHandler) doRotate() {
	if h.LogRotate > 0 && h.logFile != nil {
		baseName := path.Base(h.LogPath)
		dirName := path.Dir(h.LogPath)
		if h.LogRotate > 1 {
			for i := h.LogRotate - 1; i >= 1; i-- {
				curName := path.Join(dirName, fmt.Sprintf("%s.%d", baseName, i))
				nextName := path.Join(dirName, fmt.Sprintf("%s.%d", baseName, i+1))
				if _, err := os.Stat(curName); !os.IsNotExist(err) {
					log.Infoln("LogHandler: rotate log file ", curName, " -> ", nextName)
					if err := os.Rename(curName, nextName); err != nil {
						log.Errorln("LogHandler: could not rename log file: ", err)
					}
				}
			}
		}
		h.closeLogFile()
		if _, err := os.Stat(h.LogPath); !os.IsNotExist(err) {
			nextName := path.Join(dirName, fmt.Sprintf("%s.%d", baseName, 1))
			log.Infoln("LogHandler: rotate log file ", h.LogPath, " -> ", nextName)
			if err := os.Rename(h.LogPath, nextName); err != nil {
				log.Errorln("LogHandler: could not rename log file: ", err)
				return
			}
		}
		h.openLogFile()
	}
}

// this is a var for testing
var midnight = func() time.Duration {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(time.Hour * 24).Sub(now)
}

// Start the log handling (should NOT be run in a go routine). Reload must be called at least once
func (h *LogHandler) Start(parent context.Context) {
	h.shutdown = make(chan struct{})

	if h.DefaultWriter == nil {
		log.Fatalln("internal error: require default log writer")
	}

	log.SetOutput(h.DefaultWriter)

	if h.Debug {
		log.SetLevel(log.DebugLevel)
	} else if h.Verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	if h.LogPath != "" && h.logFile == nil {
		h.openLogFile()
	}

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ctx, cancel := context.WithCancel(parent)
		defer cancel()

		t := time.NewTimer(midnight())
		defer func() {
			// this is a function, because the pointer would be resolved statically
			// this way the pointer is resolved at the time the function will be called
			// we set a new timer after each day
			if !t.Stop() {
				<-t.C
			}
		}()

		defer h.closeLogFile()

		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-h.shutdown:
				if !ok {
					return
				}
			case <-t.C:
				h.doRotate()
				t.Stop()
				t = time.NewTimer(midnight())
			}
		}
	}()
}

// Shutdown all files
func (h *LogHandler) Shutdown() {
	close(h.shutdown)
	h.wg.Wait()
}
