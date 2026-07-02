package models

type Config struct {
	JavaHome          string `json:"java_home"`
	CurrentJDKVersion string `json:"current_jdk_version"`
	Originalpath      string `json:"original_path"`
	Proxy             string `json:"proxy"`
	Store             string
	Download          string
}
