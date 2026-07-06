package jdk

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/baneeishaque/adoptium_jdk_go"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/web"
)

var getJdkLock sync.Mutex

type fetchResult struct {
	name     string
	versions []models.JdkVersion
	err      error
	duration time.Duration
}

// GetJdkVersions acquires sync lock, fetches from remote & catches result for 24h then free lock
func GetJdkVersions(config *models.Config) ([]models.JdkVersion, error) {
	getJdkLock.Lock()
	defer getJdkLock.Unlock()

	if versions, _, err := loadCachedJdkVersions(config); err == nil {
		return versions, nil
	}

	fmt.Println("Fetching available JDK versions...")
	start := time.Now()
	results := make(chan fetchResult, 3)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		start := time.Now()
		jsonContent, err := web.GetRemoteTextFile(config.OriginalPath)
		if err != nil {
			results <- fetchResult{err: err}
			return
		}
		var versions []models.JdkVersion
		if err := json.Unmarshal([]byte(jsonContent), &versions); err != nil {
			results <- fetchResult{err: err}
			return
		}
		results <- fetchResult{
			name:     "Original",
			versions: versions,
			duration: time.Since(start),
		}
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		adoptiumJdks := strings.Split(adoptium_jdk_go.ApiListReleases(), "\n")
		versions := make([]models.JdkVersion, 0, len(adoptiumJdks))
		for _, adoptiumJdkUrl := range adoptiumJdks {
			fileSeparatorIndex := strings.LastIndex(adoptiumJdkUrl, "/")
			fileName := adoptiumJdkUrl[fileSeparatorIndex+1:]
			fileVersion := strings.TrimSuffix(fileName, ".zip")
			versions = append(versions, models.JdkVersion{
				Version: fileVersion,
				Url:     adoptiumJdkUrl,
			})
		}
		results <- fetchResult{
			name:     "Adoptium",
			versions: versions,
			duration: time.Since(start),
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()

		versions, err := getAzulJdks()
		if err != nil {
			log.Printf("could not fetch azul jdk: %v", err)
			results <- fetchResult{
				name:     "Azul",
				duration: time.Since(start),
			}
			return
		}
		results <- fetchResult{
			name:     "Azul",
			versions: versions,
			duration: time.Since(start),
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	timeout := time.After(30 * time.Second)
	var versions []models.JdkVersion

collect:
	for {
		select {
		case result, ok := <-results:
			if !ok {
				break collect
			}
			if result.err != nil {
				return nil, result.err
			}

			if result.name != "" {
				log.Printf("%s: %d versions in %s", result.name, len(result.versions), result.duration)
			}

			versions = append(versions, result.versions...)
		case <-timeout:
			return nil, fmt.Errorf("timed out fetching JDK versions")
		}
	}
	log.Printf("Fetched %d JDK versions in %s", len(versions), time.Since(start))

	// Cache the fetched versions for future use
	if err := cacheJdkVersions(config, versions); err != nil {
		return nil, err
	}

	return versions, nil
}
