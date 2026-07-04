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
	"github.com/ystyle/jvms/utils/jdk"
	"github.com/ystyle/jvms/utils/web"
)

var (
	version = "2.1.0"
	config  = models.NewConfigPtr()
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
		log.Fatal(err.Error()) // Fatal already calls os.Exit(1)
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

	go func() {
		if _, err := jdk.GetJdkVersions(config); err != nil {
			log.Printf("background JDK preload failed: %v", err)
		}
	}()
	if config.Proxy != "" {
		web.SetProxy(config.Proxy)
	}
	return nil
}

func shutdown(c *cli.Context) error {
	if err := store.Save(models.ConfigFileName, config); err != nil {
		return errors.New("failed to save the config:" + err.Error())
	}
	return nil
}

func commandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

func marshalFunc(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}
