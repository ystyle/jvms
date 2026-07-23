package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/icli"
	"github.com/ystyle/jvms/internal/models"
)

var (
	version = "2.1.0"
	config  = models.NewConfig()
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
	}
}

func startup(c *cli.Context) error {
	store.Register("json", marshalFunc, json.Unmarshal)
	store.Init(models.ProjectConfigDir)

	// Load the config and store with idempotent initialization
	if err := store.Load(models.ConfigFileName, config); err != nil {
		return errors.New("failed to load the config:" + err.Error())
	}
	config.Load() // Ensure the config is initialized with idempotence

	return nil
}

func shutdown(c *cli.Context) error {
	return config.Save()
}

func commandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

func marshalFunc(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}
