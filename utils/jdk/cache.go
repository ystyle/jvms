package jdk

import (
	"errors"
	"time"

	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/models"
)

const (
	cacheFileName = "jdk_versions.json"
	cacheTTL      = 24 * 60 * 60 // Cache time-to-live in seconds (24 hours)
)

type JdkVersionCache struct {
	Versions    []models.JdkVersion `json:"versions"`
	LastUpdated int64               `json:"last_updated"`
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

func loadCachedJdkVersions(config *models.Config) ([]models.JdkVersion, error) {
	versionsCached := &JdkVersionCache{}
	if err := store.Load(cacheFileName, versionsCached); err != nil {
		return nil, err
	}

	// Check if the cache is still valid based on TTL
	if time.Now().Unix()-versionsCached.LastUpdated > cacheTTL {
		return nil, errors.New("cached JDK versions are stale")
	}

	return versionsCached.Versions, nil
}
