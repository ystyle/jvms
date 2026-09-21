package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

func TestCommandsWriteToAppWriter(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"list"}, "1) 21-tem"},
		{[]string{"install", "21-tem"}, "already installed"},
		{[]string{"switch", "21-tem"}, "Now using JDK 21-tem"},
		{[]string{"use", "21-tem"}, "Now using JDK 21-tem"},
		{[]string{"switch", "--as_path", "C:/Java/jdk"}, "Now using JDK at C:/Java/jdk"},
		{[]string{"remove", "missing"}, "run `jvms list`"},
		{[]string{"proxy", "--show"}, "Current proxy:"},
		{[]string{"init"}, "provider is installed and working"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			manager := &recordingManager{installed: []platform.Installation{{Version: "21-tem"}}}
			var output bytes.Buffer
			app := cli.NewApp()
			app.Writer = &output
			app.Commands = Commands(appcfg.NewConfig(), manager)
			if err := app.Run(append([]string{"jvms"}, tc.args...)); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), tc.want) {
				t.Fatalf("output = %q, want %q", output.String(), tc.want)
			}
		})
	}
}

func TestRemoveRequiresVersion(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []cli.Command{*remove(&recordingManager{})}
	err := app.Run([]string{"jvms", "remove", " "})
	if err == nil || !strings.Contains(err.Error(), "a JDK version is required") {
		t.Fatalf("error = %v", err)
	}
}
