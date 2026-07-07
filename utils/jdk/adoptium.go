package jdk

import (
	"strings"

	"github.com/baneeishaque/adoptium_jdk_go"
	"github.com/ystyle/jvms/internal/models"
)

type adoptiumProvider struct{}

func (p adoptiumProvider) Name() string {
	return "Adoptium"
}

func (p adoptiumProvider) Fetch(out chan<- models.JdkVersion) error {
	adoptiumJdks := strings.Split(adoptium_jdk_go.ApiListReleases(), "\n")

	for _, adoptiumJdkUrl := range adoptiumJdks {
		if adoptiumJdkUrl == "" {
			continue
		}

		fileSeparatorIndex := strings.LastIndex(adoptiumJdkUrl, "/")
		fileName := adoptiumJdkUrl[fileSeparatorIndex+1:]
		fileVersion := strings.TrimSuffix(fileName, ".zip")

		out <- models.JdkVersion{
			Version: fileVersion,
			Url:     adoptiumJdkUrl,
		}
	}

	return nil
}