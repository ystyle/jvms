package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

func SetProxy(p string) {
	if p != "" && p != "none" {
		proxyUrl, _ := url.Parse(p)
		client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		}
	} else {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
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

func GetBytes(url string, timeout time.Duration) ([]byte, error) {
	httpClient := *client
	httpClient.Timeout = timeout

	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("%s: %s", response.Status, body)
	}

	return io.ReadAll(response.Body)
}

func GetRemoteTextFile(url string) (string, error) {
	contents, err := GetBytes(url, 0)
	if err != nil {
		return "", fmt.Errorf("\nCould not retrieve %s.\n\n%s\n", url, err.Error())
	}
	return string(contents), nil
}
