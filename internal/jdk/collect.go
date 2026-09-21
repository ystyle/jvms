package jdk

import (
	"fmt"
	"log"
	"time"
)

func collectProviderVersions(
	versionsOut <-chan JdkVersion,
	results <-chan providerResult,
	mute bool,
	sink func(JdkVersion),
) ([]JdkVersion, []error, error) {
	timeout := time.After(providerFetchTimeout)
	var versions []JdkVersion
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
			return versions, errs, fmt.Errorf("timed out fetching JDK versions")
		}
	}

	return versions, errs, nil
}
