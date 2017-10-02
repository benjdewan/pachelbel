package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deprovisionCmd = &cobra.Command{
	Use:   "deprovision",
	Short: "Idempotent deprovisioner of compose deployments",
	Long: `pachelbel deprovision reads a mixed list of deployment names and/or IDs as
argumentsr For each deployment pachelbel will, by default, and then makes
deprovisioning requests to the Compose API. If a specified deployment does not
exist or has already been deleted, it is skipped.`,
	Run: runDeprovision,
}

func runDeprovision(cmd *cobra.Command, args []string) {
	assertCanDeprovision(args)
	cxn, err := connection.New(viper.GetString("log-file"),
		viper.GetBool("dry-run"))
	if err != nil {
		log.Fatal(err)
	}

	if err := cxn.Init(viper.GetString("api-key")); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := cxn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()
	deprovision(cxn, args)
}

func deprovision(cxn *connection.Connection, deployments []string) {
	timeout := float64(viper.GetInt("timeout"))
	if !viper.GetBool("wait") {
		timeout = 0
	}
	if err := cxn.Deprovision(deployments, timeout); err != nil {
		log.Fatal(err)
	}
}

func init() {
	deprovisionCmd.Flags().BoolP("wait", "w", false,
		`Wait for deprovisioning recipes to complete before
			returning`)
	viper.BindPFlag("wait", deprovisionCmd.Flags().Lookup("wait"))
	deprovisionCmd.Flags().IntP("timeout", "t", 300,
		`The amount of time to wait, in seconds, for
			deprovisioning recipes to complete.

			Ignored if '--wait' is not set`)
	viper.BindPFlag("timeout", deprovisionCmd.Flags().Lookup("timeout"))

	RootCmd.AddCommand(deprovisionCmd)
}

func assertCanDeprovision(args []string) {
	if len(args) == 0 {
		fmt.Println("Nothing to do")
		os.Exit(0)
	}
}
