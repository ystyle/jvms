package jdk

import (
	"log"

	"github.com/ystyle/jvms/internal/models"
)

const versionEventBuffer = 256

type VersionEvent struct {
	Version models.JdkVersion
	Err     error
	Done    bool
}

func StreamJdkVersions(config *models.Config) <-chan VersionEvent {
	events := make(chan VersionEvent, versionEventBuffer)

	go func() {
		defer close(events)
		streamJdkVersions(config, events)
	}()

	return events
}

func streamJdkVersions(config *models.Config, events chan<- VersionEvent) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	versions, err := loadCachedJdkVersions()
	if err == nil {
		sendCachedVersions(versions, events)
		return
	}

	versions, errs, err := fetchJdkVersionsWithSink(configuredProviders(config), true, func(version models.JdkVersion) {
		events <- VersionEvent{Version: version}
	})
	if err != nil {
		log.Printf("fetch timed out (partial results: %d versions): %v", len(versions), err)
	} else if len(errs) > 0 {
		log.Printf("Fetched JDK versions with %d provider error(s)", len(errs))
	}

	if err := cacheJdkVersions(versions); err != nil {
		events <- VersionEvent{Err: err, Done: true}
		return
	}
	events <- VersionEvent{Done: true}
}

func sendCachedVersions(versions []models.JdkVersion, events chan<- VersionEvent) {
	for _, version := range versions {
		events <- VersionEvent{Version: version}
	}
	events <- VersionEvent{Done: true}
}
