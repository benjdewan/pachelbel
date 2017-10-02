package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "DEV-BUILD"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "The version of pachelbel",
	Long:  "The version of pachelbel",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
