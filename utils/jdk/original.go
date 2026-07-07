package jdk

import (
	"encoding/json"

	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/web"
)

type originalProvider struct {
	url string
}

func (p originalProvider) Name() string {
	return "Original"
}

func (p originalProvider) Fetch(out chan<- models.JdkVersion) error {
	jsonContent, err := web.GetRemoteTextFile(p.url)
	if err != nil {
		return err
	}

	var versions []models.JdkVersion
	if err := json.Unmarshal([]byte(jsonContent), &versions); err != nil {
		return err
	}

	for _, version := range versions {
		out <- version
	}

	return nil
}