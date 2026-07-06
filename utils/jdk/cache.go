package jdk

import (
	"errors"
	"time"

	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/models"
)

const (
	cacheFileName    = "jdk_versions.json"
	cacheTTL         = 24 * 60 * 60 // Cache time-to-live in seconds (24 hours)
	cacheFetchWindow = 30
)

type JdkVersionCache struct {
	Versions    []models.JdkVersion `json:"versions"`
	LastUpdated int64               `json:"last_updated"`
}

func InvalidateCache(config *models.Config) error {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	_, recentlyFetched, err := loadCachedJdkVersions(config)
	if err == nil && recentlyFetched {
		// Cache was refreshed very recently; don't throw it away.
		return nil
	}

	return store.Save(cacheFileName, &JdkVersionCache{})
}

func cacheJdkVersions(config *models.Config, versions []models.JdkVersion) error {
	if len(versions) == 0 {
		return errors.New("no JDK versions to cache")
	}

	return store.Save(cacheFileName, &JdkVersionCache{
		Versions:    versions,
		LastUpdated: time.Now().Unix(),
	})
}

// loadCachedJdkVersions returns the cached versions and whether the cache was
// refreshed recently.
func loadCachedJdkVersions(config *models.Config) ([]models.JdkVersion, bool, error) {
	versionsCached := &JdkVersionCache{}
	if err := store.Load(cacheFileName, versionsCached); err != nil {
		return nil, false, err
	}

	age := time.Now().Unix() - versionsCached.LastUpdated

	if age > cacheTTL {
		return nil, age <= cacheFetchWindow, errors.New("cached JDK versions are stale")
	}

	return versionsCached.Versions, age <= cacheFetchWindow, nil
}
