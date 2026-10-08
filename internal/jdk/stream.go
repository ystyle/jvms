package jdk

import (
	"log"

	appcfg "github.com/ystyle/jvms/internal/config"
)

const versionEventBuffer = 256

type VersionEvent struct {
	Version JdkVersion
	Err     error
	Done    bool
}

func StreamJdkVersions(config *appcfg.Config) <-chan VersionEvent {
	events := make(chan VersionEvent, versionEventBuffer)

	go func() {
		defer close(events)
		streamJdkVersions(config, events)
	}()

	return events
}

func streamJdkVersions(config *appcfg.Config, events chan<- VersionEvent) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	versions, err := LoadCachedVersions(config, nativeCacheSource, false)
	if err == nil {
		sendCachedVersions(versions, events)
		return
	}

	versions, errs, err := fetchJdkVersionsWithSink(configuredProviders(config), true, func(version JdkVersion) {
		events <- VersionEvent{Version: version}
	})
	if err != nil {
		log.Printf("fetch timed out (partial results: %d versions): %v", len(versions), err)
	} else if len(errs) > 0 {
		log.Printf("Fetched JDK versions with %d provider error(s)", len(errs))
	}

	if err := CacheVersions(config, nativeCacheSource, versions); err != nil {
		events <- VersionEvent{Err: err, Done: true}
		return
	}

	events <- VersionEvent{Done: true}
}

func sendCachedVersions(versions []JdkVersion, events chan<- VersionEvent) {
	for _, version := range versions {
		events <- VersionEvent{Version: version}
	}

	events <- VersionEvent{Done: true}
}
