package jdk

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ystyle/jvms/internal/models"
)

var getJdkLock sync.Mutex

// GetJdkVersions returns cached versions when available, otherwise it fetches
// from remote providers and refreshes the cache.
func GetJdkVersions(config *models.Config, mute bool) ([]models.JdkVersion, error) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	if versions, err := loadCachedJdkVersions(); err == nil {
		return versions, nil
	}

	fmt.Println("Fetching available JDK versions...")

	start := time.Now()
	versions, errs, err := fetchJdkVersions(configuredProviders(config), mute)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		log.Printf("Fetched JDK versions with %d provider error(s)", len(errs))
		return versions, nil
	}

	log.Printf("Fetched %d JDK versions in %s", len(versions), time.Since(start))

	if err := cacheJdkVersions(versions); err != nil {
		return nil, err
	}

	return versions, nil
}
