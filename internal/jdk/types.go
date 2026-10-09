package jdk

type JdkVersion struct {
	Version string `json:"version"`
	Url     string `json:"url"`
}

type Version = JdkVersion

type Installation struct {
	Version string
	Current bool
}
