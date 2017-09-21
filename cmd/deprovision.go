// Copyright © 2017 ben dewan <benj.dewan@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
