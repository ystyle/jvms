package icli

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

func rls(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:  "rls",
		Usage: "Show a list of versions available for download. ",
		Flags: []cli.Flag{cli.BoolFlag{Name: "a", Usage: "list all the version"}},
		Action: func(c *cli.Context) error {
			if err := jdk.InvalidateCache(); err != nil {
				log.Printf("failed to invalidate cache: %v", err)
			} // No more race

			versions, err := jdk.GetJdkVersions(config, true)
			if err != nil {
				return err
			}
			for i, version := range versions {
				fmt.Printf("    %d) %s\n", i+1, version.Version)
				if !c.Bool("a") && i >= 9 {
					fmt.Println("\nuse \"jvm rls -a\" show all the versions ")
					break
				}
			}
			if len(versions) == 0 {
				fmt.Println("No available jdk version for download.")
			}

			fmt.Printf("\nFor a complete list, visit %s\n", config.OriginalPath)
			return nil
		},
	}
}
