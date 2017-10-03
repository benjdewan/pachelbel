package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deprovisionCmd = &cobra.Command{
	Use:   "deprovision",
	Short: "Idempotent deprovisioner of compose deployments",
	Long: `pachelbel deprovision reads a list of deployment names as
arguments. For each deployment pachelbel will, check if the deployment exists,
and if it does sends a deprovisioning request to the Compose API. If a
specified deployment does not exist or has already been deleted it is ignored.`,
	Run: runDeprovision,
}

func runDeprovision(cmd *cobra.Command, args []string) {
	assertCanDeprovision(args)
	file, err := writeDeprovisionFile(args)
	if err != nil {
		log.Fatal(err)
	}
	provisionCmd.Run(cmd, []string{file})
	if err = os.Remove(file); err != nil {
		log.Fatal(err)
	}
}

const deprovisionTemplate = `config_version: 2
object_type: deprovision
name: %s
timeout: %d`

func writeDeprovisionFile(args []string) (string, error) {
	timeout := viper.GetInt("timeout")
	if !viper.GetBool("wait") {
		timeout = 0
	}
	objs := []string{}
	for _, arg := range args {
		objs = append(objs, fmt.Sprintf(deprovisionTemplate, arg, timeout))
	}
	handle, err := ioutil.TempFile("", "deprovision")
	if err != nil {
		return "", err
	}
	_, err = handle.WriteString(strings.Join(objs, "\n---\n"))
	if err != nil {
		return "", err
	}
	return handle.Name(), handle.Close()
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
