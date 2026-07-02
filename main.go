package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/internal/icli"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/web"
)

var (
	version = "2.1.0"
	config  = &entity.Config{}
)

func main() {
	app := cli.NewApp()
	app.Name = "jvms"
	app.Usage = `JDK Version Manager (JVMS) for Windows`
	app.Version = version
	app.CommandNotFound = commandNotFound
	app.Commands = icli.Commands(config)

	app.Before = startup
	app.After = shutdown
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
}

func commandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

func startup(c *cli.Context) error {
	store.Register(
		"json",
		func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		},
		json.Unmarshal)

	store.Init("jvms")
	if err := store.Load("jvms.json", &config); err != nil {
		return errors.New("failed to load the config:" + err.Error())
	}
	s := file.GetCurrentPath()

	config.Store = filepath.Join(s, "store")
	// Override store path by storepath file
	if storepath, err := os.ReadFile(filepath.Join(s, "storepath")); err == nil {
		config.Store = string(storepath)
	}
	config.Download = filepath.Join(s, "download")
	if config.Originalpath == "" {
		config.Originalpath = icli.DefaultOriginalpath
	}
	if config.Proxy != "" {
		web.SetProxy(config.Proxy)
	}
	return nil
}

func shutdown(c *cli.Context) error {
	if err := store.Save("jvms.json", &config); err != nil {
		return errors.New("failed to save the config:" + err.Error())
	}
	return nil
}
