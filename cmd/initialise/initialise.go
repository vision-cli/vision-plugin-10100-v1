package initialise

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	api "github.com/vision-cli/api/v1"
	"github.com/vision-cli/vision-plugin-10100-v1/cmd/model"
)

var InitCmd = &cobra.Command{
	Use:   "init [app golang module in the for github.com/org/app...]",
	Short: "create a 10100 app",
	Long:  "create a 10100 app with the given golang module name",
	Args: func(cmd *cobra.Command, args []string) error {

		for _, arg := range args {
			if !strings.HasPrefix(arg, "github.com/") && strings.Count(arg, "/") > 2 {
				return fmt.Errorf("%s must be in the form github.com/org/app...", arg)
			}
		}

		return nil
	},
	RunE: runCommand,
}

func runCommand(cmd *cobra.Command, args []string) error {
	jEnc := json.NewEncoder(os.Stdout)

	pd := model.PluginConfig{
		Modules: args,
	}

	err := jEnc.Encode(api.Init{
		Config:  pd,
		Success: true,
	})

	if err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}

	err = installTempl()
	if err != nil {
		return fmt.Errorf("insalling templ: %w", err)
	}

	return nil
}

func installTempl() error {
	c := exec.Command("go", "install", "github.com/a-h/templ/cmd/templ@latest")
	_, err := c.Output()
	if err != nil {
		return fmt.Errorf("running 'go install templ': %w", err)
	}
	return nil
}
