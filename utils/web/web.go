package web

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"io"
	"io/ioutil"
	"errors"
)

var client = &http.Client{}

func SetProxy(p string) {
	if p != "" && p != "none" {
		proxyUrl, _ := url.Parse(p)
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	} else {
		client = &http.Client{}
	}
}

func Download(url string, target string) bool {
	output, err := os.Create(target)
	if err != nil {
		fmt.Println("Error while creating", target, "-", err)
	}
	defer output.Close()

	response, err := client.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
	}

	if response.Status[0:3] != "200" {
		fmt.Println("Download failed. Rolling Back.")
		err := os.Remove(target)
		if err != nil {
			fmt.Println("Rollback failed.", err)
		}
		return false
	}

	return true
}

func GetJDK(download string, v string, url string) bool {

	if url == "" {
		//No url should mean this version/arch isn't available
		fmt.Printf("JDK %s isn't available right now.", v)
	} else {
		fileName := fmt.Sprintf("%s%s.zip", download, v)
		fmt.Printf("Downloading jdk version %s...\n", v)
		if Download(url, fileName) {
			fmt.Println("Complete")
			return true
		} else {
			return false
		}
	}
	return false

}

func GetRemoteTextFile(url string) (string,error) {
	response, httperr := client.Get(url)
	if httperr != nil {
		return "", errors.New(fmt.Sprintf("\nCould not retrieve %s.\n\n%s\n",url,httperr.Error()))
	} else {
		defer response.Body.Close()
		contents, readerr := ioutil.ReadAll(response.Body)
		if readerr != nil {
			return "", errors.New(fmt.Sprintf("%s", readerr))
		}
		return string(contents),nil
	}
}
