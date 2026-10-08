package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	appcli "github.com/ystyle/jvms/internal/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

var (
	version = "2.1.0"
	config  = appcfg.NewConfig()
)

func main() {
	app := cli.NewApp()
	app.Name = "jvms"
	app.Usage = `JDK Version Manager (JVMS) for Windows`
	app.Version = version
	app.CommandNotFound = appcli.CommandNotFound
	app.Commands = appcli.Commands(config, platform.NewProvider(config))

	app.Before = func(c *cli.Context) error { return config.Load() }
	app.After = func(c *cli.Context) error { return config.Save() }

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
