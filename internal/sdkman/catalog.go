//go:build !windows

package sdkman

import (
	"errors"
	"fmt"

	"github.com/ystyle/jvms/internal/jdk"
)

const sdkmanCacheSource = "sdkman"

func (m *service) RefreshAvailable() ([]jdk.Version, error) {
	return m.fetchAvailable()
}

func (m *service) Available() ([]jdk.Version, error) {
	if versions, err := jdk.LoadCachedVersions(m.config, sdkmanCacheSource, false); err == nil {
		return versions, nil
	}

	return m.fetchAvailable()
}

// Keep the previous catalog available until SDKMAN returns a usable replacement.
func (m *service) fetchAvailable() ([]jdk.Version, error) {
	output, err := m.run(false, "list", "java")
	if err != nil {
		return m.cachedSDKMANVersions(err)
	}

	versions := parseSDKMANJavaList(output)
	if err := sdkmanListError(output); err != nil {
		if cached, cacheErr := m.cachedSDKMANVersions(err); cacheErr == nil {
			return cached, nil
		}

		return versions, err
	}

	if len(versions) == 0 {
		return m.cachedSDKMANVersions(errors.New("SDKMAN returned no available Java versions"))
	}

	if err := jdk.CacheVersions(m.config, sdkmanCacheSource, versions); err != nil {
		return versions, fmt.Errorf("cache SDKMAN Java versions: %w", err)
	}

	return versions, nil
}

// StreamAvailable emits SDKMAN's complete catalog one entry at a time.
func (m *service) StreamAvailable() <-chan jdk.VersionEvent {
	events := make(chan jdk.VersionEvent)

	go func() {
		defer close(events)

		versions, err := m.Available()
		for _, version := range versions {
			events <- jdk.VersionEvent{Version: version}
		}

		events <- jdk.VersionEvent{Err: err, Done: true}
	}()

	return events
}

func (m *service) cachedSDKMANVersions(cause error) ([]jdk.Version, error) {
	versions, err := jdk.LoadCachedVersions(m.config, sdkmanCacheSource, true)
	if err == nil {
		return versions, nil
	}

	return nil, cause
}
