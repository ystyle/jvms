package jdk

import "github.com/ystyle/jvms/internal/models"

type jdkProvider interface {
	Name() string
	Fetch(chan<- models.JdkVersion) error
}
