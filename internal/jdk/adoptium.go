package jdk

import (
	"strings"

	"github.com/baneeishaque/adoptium_jdk_go"
)

type adoptiumProvider struct{}

func (p adoptiumProvider) Name() string {
	return "Adoptium"
}

func (p adoptiumProvider) Fetch(out chan<- JdkVersion) error {
	adoptiumJdks := strings.Split(adoptium_jdk_go.ApiListReleases(), "\n")

	for _, adoptiumJdkUrl := range adoptiumJdks {
		if adoptiumJdkUrl == "" {
			continue
		}

		fileSeparatorIndex := strings.LastIndex(adoptiumJdkUrl, "/")
		fileName := adoptiumJdkUrl[fileSeparatorIndex+1:]
		fileVersion := strings.TrimSuffix(fileName, ".zip")

		out <- JdkVersion{
			Version: fileVersion,
			Url:     adoptiumJdkUrl,
		}
	}

	return nil
}
