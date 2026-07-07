package jdk

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ystyle/jvms/internal/models"
)

var getJdkLock sync.Mutex

type providerResult struct {
	name     string
	count    int
	err      error
	duration time.Duration
}

var errs []error

// GetJdkVersions acquires sync lock, fetches from remote & catches result for 24h then free lock
func GetJdkVersions(config *models.Config, mute bool) ([]models.JdkVersion, error) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	if versions, _, err := loadCachedJdkVersions(config); err == nil {
		return versions, nil
	}

	fmt.Println("Fetching available JDK versions...")

	start := time.Now()
	providers := []jdkProvider{
		azulProvider{},
		originalProvider{url: config.OriginalPath},
		adoptiumProvider{},
	}

	versionsOut := make(chan models.JdkVersion)
	results := make(chan providerResult, len(providers))

	var wg sync.WaitGroup
	wg.Add(len(providers))

	for _, provider := range providers {
		go func(provider jdkProvider) {
			defer wg.Done()

			start := time.Now()
			count := 0
			out := make(chan models.JdkVersion)

			done := make(chan error, 1)
			go func() {
				done <- provider.Fetch(out)
				close(out)
			}()

			for version := range out {
				count++
				versionsOut <- version
			}

			results <- providerResult{
				name:     provider.Name(),
				count:    count,
				err:      <-done,
				duration: time.Since(start),
			}
		}(provider)
	}

	go func() {
		wg.Wait()
		close(versionsOut)
		close(results)
	}()

	timeout := time.After(30 * time.Second)
	var versions []models.JdkVersion

	for versionsOut != nil || results != nil {
		select {
		case version, ok := <-versionsOut:
			if !ok {
				versionsOut = nil
				continue
			}
			if !mute {
				log.Printf("Got %s in %s", version.Version, time.Since(start))
			}
			versions = append(versions, version)
		case result, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			if result.err != nil {
				log.Printf("%s provider failed: %v", result.name, result.err)
				errs = append(errs, result.err)
			}
		case <-timeout:
			return nil, fmt.Errorf("timed out fetching JDK versions")
		}
	}

	log.Printf("Fetched %d JDK versions in %s", len(versions), time.Since(start))

	if err := cacheJdkVersions(config, versions); err != nil {
		return nil, err
	}

	return versions, nil
}
