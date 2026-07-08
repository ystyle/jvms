package jdk

import (
	"fmt"
	"log"
	"time"

	"github.com/ystyle/jvms/internal/models"
)

func collectProviderVersions(
	versionsOut <-chan models.JdkVersion,
	results <-chan providerResult,
	mute bool,
	sink func(models.JdkVersion),
) ([]models.JdkVersion, []error, error) {
	timeout := time.After(providerFetchTimeout)
	var versions []models.JdkVersion
	var errs []error

	for versionsOut != nil || results != nil {
		select {
		case version, ok := <-versionsOut:
			if !ok {
				versionsOut = nil
				continue
			}
			versions = append(versions, version)
			if sink != nil {
				sink(version)
			}
		case result, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			if result.err != nil {
				if !mute {
					log.Printf("%s provider failed: %v", result.name, result.err)
				}
				errs = append(errs, result.err)
			}
		case <-timeout:
			return nil, errs, fmt.Errorf("timed out fetching JDK versions")
		}
	}

	return versions, errs, nil
}
