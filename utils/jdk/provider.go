package jdk

import "github.com/ystyle/jvms/internal/models"

type jdkProvider interface {
	Name() string

	// Fetch runs the provider's single processing loop and sends each version to
	// out as soon as it is discovered. Fetch must not close out.
	Fetch(chan<- models.JdkVersion) error
}

func configuredProviders(config *models.Config) []jdkProvider {
	return []jdkProvider{
		azulProvider{},
		originalProvider{url: config.OriginalPath},
		adoptiumProvider{},
	}
}
