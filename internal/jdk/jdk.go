package jdk

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	appcfg "github.com/ystyle/jvms/internal/config"
)

var getJdkLock sync.Mutex

const nativeCacheSource = "native-windows"

// GetJdkVersions returns cached versions or refreshes them from remote providers.
func GetJdkVersions(config *appcfg.Config, mute bool) ([]Version, error) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	if versions, err := LoadCachedVersions(config, nativeCacheSource, false); err == nil {
		sortCatalog(versions)
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

	if err := CacheVersions(config, nativeCacheSource, versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// sortCatalog keeps CLI indexes independent of provider response order
func sortCatalog(versions []Version) {
	slices.SortFunc(versions, func(a, b Version) int {
		if order := cmp.Compare(b.Version, a.Version); order != 0 {
			return order
		}
		return cmp.Compare(a.Url, b.Url)
	})
}
