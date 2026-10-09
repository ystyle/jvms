package jdk

import appcfg "github.com/ystyle/jvms/internal/config"

type jdkProvider interface {
	Name() string

	// Fetch runs the provider's single processing loop and sends each version to
	// out as soon as it is discovered. Fetch must not close out.
	Fetch(chan<- JdkVersion) error
}

func configuredProviders(config *appcfg.Config) []jdkProvider {
	return []jdkProvider{
		azulProvider{},
		originalProvider{url: config.OriginalPath},
		adoptiumProvider{},
	}
}
