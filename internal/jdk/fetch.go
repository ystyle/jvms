package jdk

import (
	"sync"
	"time"
)

const providerFetchTimeout = 30 * time.Second

type providerResult struct {
	name string
	err  error
}

func fetchJdkVersions(providers []jdkProvider, mute bool) ([]Version, []error, error) {
	versions, errs, err := fetchJdkVersionsWithSink(providers, mute, nil)
	sortCatalog(versions)
	return versions, errs, err
}

func fetchJdkVersionsWithSink(
	providers []jdkProvider,
	mute bool,
	sink func(Version),
) ([]Version, []error, error) {
	versionsOut := make(chan Version)
	results := make(chan providerResult, len(providers))

	var wg sync.WaitGroup
	wg.Add(len(providers))

	for _, provider := range providers {
		go runProvider(provider, versionsOut, results, &wg)
	}

	go func() {
		wg.Wait()
		close(versionsOut)
		close(results)
	}()

	return collectProviderVersions(versionsOut, results, mute, sink)
}

func runProvider(
	provider jdkProvider,
	versionsOut chan<- Version,
	results chan<- providerResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	err := provider.Fetch(versionsOut)
	results <- providerResult{
		name: provider.Name(),
		err:  err,
	}
}
