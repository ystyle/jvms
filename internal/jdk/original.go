package jdk

import (
	"encoding/json"

	"github.com/ystyle/jvms/internal/httpclient"
)

type originalProvider struct {
	url string
}

func (p originalProvider) Name() string {
	return "Original"
}

func (p originalProvider) Fetch(out chan<- JdkVersion) error {
	jsonContent, err := httpclient.GetRemoteTextFile(p.url)
	if err != nil {
		return err
	}

	var versions []JdkVersion
	if err := json.Unmarshal([]byte(jsonContent), &versions); err != nil {
		return err
	}

	for _, version := range versions {
		out <- version
	}

	return nil
}
