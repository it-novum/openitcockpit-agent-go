package platformpaths

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// PlatformPath represents the default paths for specific directory and files
type PlatformPath interface {
	// Init will be automatically called by Get()
	Init() error
	// LogPath for this system
	LogPath() string
	// ConfigPath for this system
	ConfigPath() string
	// AdditionalData for this plattform (only on windows)
	AdditionalData() map[string]string
}

var (
	mtx      sync.Mutex
	instance PlatformPath
)

// Get the platform path instance
func Get() PlatformPath {
	mtx.Lock()
	defer mtx.Unlock()
	if instance == nil {
		instance = getPlatformPath()
		if err := instance.Init(); err != nil {
			log.Errorln(err)
		}
	}
	return instance
}
