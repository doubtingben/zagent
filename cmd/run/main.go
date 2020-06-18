package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RunCmd runs an instance of zcashd
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an instance of zcashd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running zcashd")

	},
}
