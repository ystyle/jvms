package jdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/ystyle/jvms/internal/models"
)

type azulProvider struct{}

type azulJdk struct {
	PackageUUID        string `json:"package_uuid"`
	Name               string `json:"name"`
	JavaVersion        []int  `json:"java_version"`
	OpenjdkBuildNumber int    `json:"openjdk_build_number"`
	Latest             bool   `json:"latest"`
	DownloadURL        string `json:"download_url"`
	Product            string `json:"product"`
	DistroVersion      []int  `json:"distro_version"`
	AvailabilityType   string `json:"availability_type"`
	ShortName          string
}

const azulApi = "https://api.azul.com/metadata/v1/zulu/packages"

func (p azulProvider) Name() string {
	return "Azul"
}

func (p azulProvider) Fetch(out chan<- models.JdkVersion) error {
	body, err := call(AzulApiEndpoint())
	if err != nil {
		return fmt.Errorf("error %v", err)
	}

	var jdks []azulJdk
	if err := json.Unmarshal(body, &jdks); err != nil {
		return fmt.Errorf("error %v", err)
	}

	for _, jdk := range jdks {
		lastIndex := strings.LastIndex(jdk.Name, "-")
		if lastIndex <= 0 || jdk.DownloadURL == "" {
			continue
		}

		out <- models.JdkVersion{
			Version: jdk.Name[:lastIndex],
			Url:     jdk.DownloadURL,
		}
	}

	return nil
}

func AzulApiEndpoint() string { //https://api.azul.com/metadata/v1/docs/swagger
	var api = azulApi + "?os=$OS&arch=$ARCH&archive_type=zip&java_package_type=jdk&javafx_bundled=false&latest=true&release_status=ga&availability_types=CA&certifications=tck&page=1&page_size=100"
	api = strings.Replace(api, "$OS", runtime.GOOS, 1)
	api = strings.Replace(api, "$ARCH", runtime.GOARCH, 1)
	return api
}

func call(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("%s: %s", res.Status, body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
