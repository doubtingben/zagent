package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "v0.0.0.0-dev"
	GoVersion = runtime.Version()
	GitCommit string
	GitBranch string
	BuildDate string
	BuildUser string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Dispaly version",
	Long:  `Dispaly version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", Version)
		fmt.Println("from commit: ", GitCommit)
		fmt.Println("on: ", BuildDate)
		fmt.Println("by: ", BuildUser)
	},
}
