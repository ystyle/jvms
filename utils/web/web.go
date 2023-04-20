package web

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
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
	response, err := client.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return false
	}
	if response.StatusCode != 200 {
		fmt.Println("Error status while downloading", url, "-", response.StatusCode)
		return false
	}
	defer response.Body.Close()

	output, err := os.Create(target)
	if err != nil {
		fmt.Println("Error while creating", target, "-", err)
		return false
	}
	defer output.Close()

	// 创建一个进度条
	bar := pb.New(int(response.ContentLength)).SetUnits(pb.U_BYTES_DEC).SetRefreshRate(time.Millisecond * 10)
	// 显示下载速度
	bar.ShowSpeed = true

	// 显示剩余时间
	bar.ShowTimeLeft = true

	// 显示完成时间
	bar.ShowFinalTime = true

	bar.SetWidth(80)

	bar.Start()
	writer := io.MultiWriter(output, bar)
	_, err = io.Copy(writer, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return false
	}
	bar.Finish()

	return true
}

func GetJDK(download string, v string, url string) (string, bool) {
	fileName := filepath.Join(download, fmt.Sprintf("%s.zip", v))
	os.Remove(fileName)
	if url == "" {
		//No url should mean this version/arch isn't available
		fmt.Printf("JDK %s isn't available right now.", v)
	} else {
		fmt.Printf("Downloading jdk version %s...\n", v)
		if Download(url, fileName) {
			fmt.Println("Complete")
			return fileName, true
		} else {
			return "", false
		}
	}
	return "", false

}

func GetRemoteTextFile(url string) (string, error) {
	response, httperr := client.Get(url)
	if httperr != nil {
		return "", errors.New(fmt.Sprintf("\nCould not retrieve %s.\n\n%s\n", url, httperr.Error()))
	} else {
		defer response.Body.Close()
		contents, readerr := ioutil.ReadAll(response.Body)
		if readerr != nil {
			return "", errors.New(fmt.Sprintf("%s", readerr))
		}
		return string(contents), nil
	}
}
