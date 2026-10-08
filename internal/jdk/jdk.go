package jdk

import (
	"cmp"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	appcfg "github.com/ystyle/jvms/internal/config"
)

var getJdkLock sync.Mutex

const nativeCacheSource = "native-windows"

var catalogVersionNumbers = regexp.MustCompile(`[0-9]+`)

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
		if order := compareCatalogVersions(b.Version, a.Version); order != 0 {
			return order
		}
		if order := cmp.Compare(b.Version, a.Version); order != 0 {
			return order
		}
		return cmp.Compare(a.Url, b.Url)
	})
}

// Compare digit groups in order so both major and patch numbers sort numerically.
func compareCatalogVersions(a, b string) int {
	aNumbers := catalogVersionNumbers.FindAllString(a, -1)
	bNumbers := catalogVersionNumbers.FindAllString(b, -1)

	return slices.CompareFunc(aNumbers, bNumbers, func(a, b string) int {
		aNumber, _ := strconv.Atoi(a)
		bNumber, _ := strconv.Atoi(b)
		return cmp.Compare(aNumber, bNumber)
	})
}
