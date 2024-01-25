package version

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	api "github.com/vision-cli/api/v1"
	"github.com/vision-cli/vision-plugin-10100-v1/cmd/model"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "plugin version",
	Long:  "semantic version of the plugin",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		json.NewEncoder(os.Stdout).Encode(
			api.Version{
				SemVer: model.PluginSemVer,
			})
	},
}
