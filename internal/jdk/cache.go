package jdk

import (
	"errors"
	"sync"
	"time"

	"github.com/tucnak/store"
	appcfg "github.com/ystyle/jvms/internal/config"
)

const cacheFileName = "jdk_versions.json"

var cacheLock sync.Mutex

type JdkVersionCache struct {
	Versions    []Version `json:"versions"`
	LastUpdated int64     `json:"last_updated"`
	Source      string    `json:"source,omitempty"`
}

func InvalidateCache() error {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	store.Init(appcfg.ProjectConfigDir)
	return store.Save(cacheFileName, &JdkVersionCache{})
}

func CacheVersions(config *appcfg.Config, source string, versions []Version) error {
	if config != nil && !config.CacheEnabled {
		return nil
	}
	if len(versions) == 0 {
		return errors.New("no JDK versions to cache")
	}

	cacheLock.Lock()
	defer cacheLock.Unlock()
	store.Init(appcfg.ProjectConfigDir)
	return store.Save(cacheFileName, &JdkVersionCache{
		Versions:    versions,
		LastUpdated: time.Now().Unix(),
		Source:      source,
	})
}

// LoadCachedVersions optionally accepts stale entries for offline fallback.
func LoadCachedVersions(config *appcfg.Config, source string, allowStale bool) ([]Version, error) {
	if config != nil && !config.CacheEnabled {
		return nil, errors.New("JDK version cache is disabled")
	}
	cacheLock.Lock()
	defer cacheLock.Unlock()
	store.Init(appcfg.ProjectConfigDir)
	versionsCached := &JdkVersionCache{}
	if err := store.Load(cacheFileName, versionsCached); err != nil {
		return nil, err
	}

	if len(versionsCached.Versions) == 0 {
		return nil, errors.New("cached JDK versions are empty")
	}
	if versionsCached.Source != source && !(versionsCached.Source == "" && source == nativeCacheSource) {
		return nil, errors.New("cached JDK versions belong to another provider")
	}
	age := time.Since(time.Unix(versionsCached.LastUpdated, 0))
	if !allowStale && age > config.CacheDuration() {
		return nil, errors.New("cached JDK versions are stale")
	}

	return versionsCached.Versions, nil
}
