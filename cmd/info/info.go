package info

import (
	_ "embed"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	api "github.com/vision-cli/api/v1"
)

//go:embed info.txt
var infoOutput string

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "return info about the plugin",
	Long:  "return detailed information about the plugin",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		json.NewEncoder(os.Stdout).Encode(api.Info{
			ShortDescription: "This plugin creates a starter 10100 app",
			LongDescription:  infoOutput,
		})
	},
}
