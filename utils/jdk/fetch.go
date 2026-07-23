package jdk

import (
	"sync"
	"time"

	"github.com/ystyle/jvms/internal/models"
)

const providerFetchTimeout = 30 * time.Second

type providerResult struct {
	name string
	err  error
}

func fetchJdkVersions(providers []jdkProvider, mute bool) ([]models.JdkVersion, []error, error) {
	return fetchJdkVersionsWithSink(providers, mute, nil)
}

func fetchJdkVersionsWithSink(
	providers []jdkProvider,
	mute bool,
	sink func(models.JdkVersion),
) ([]models.JdkVersion, []error, error) {
	versionsOut := make(chan models.JdkVersion)
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
	versionsOut chan<- models.JdkVersion,
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
