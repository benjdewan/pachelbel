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
	"log"

	"github.com/benjdewan/pachelbel/config"
	"github.com/benjdewan/pachelbel/connection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// provisionCmd represents the provision command
var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Idempotent provisioner of compose deployments",
	Long: `pachelbel provision reads in YAML configuration(s) describing a list of
deployments that should exist in a list of clusters of specified sizes,
and ensures they do.

If the deployments do not exist, they are created. If they exist, but are
the wrong size they are scaled. If they are deployed as specified in the
configuration no actions are taken.`,
	Run: doProvision,
}

func doProvision(cmd *cobra.Command, args []string) {
	config.BuildClusterFilter(viper.GetStringSlice("cluster"))

	verbose := viper.GetBool("verbose")
	deployments, err := config.ReadFiles(args, verbose)
	if err != nil {
		log.Fatal(err)
	}

	cxn, err := connection.Init(viper.GetString("api-key"), verbose)
	if err != nil {
		log.Fatal(err)
	}

	failFast := viper.GetBool("fail-fast")
	for _, deployment := range deployments {
		err = connection.Provision(cxn, deployment, verbose)
		if err != nil {
			if failFast {
				log.Fatal(err)
			}
			log.Printf("%v. Continuing...", err)
		}
	}
}

func init() {
	RootCmd.AddCommand(provisionCmd)
	provisionCmd.Flags().BoolP("fail-fast", "f", false,
		`By default pachelbel provision does not stop until it has processed
			every deployment provided. With --fail-fast set the
			first deployment that fails will terminate
			processing.`)
	provisionCmd.Flags().StringSliceP("cluster", "c", []string{},
		`By default pachelbel provision will provision every deployment
			provided. Use this flag to limit pachelbel to only
			process deployments to the specified cluster`)

	viper.BindPFlag("fail-fast", provisionCmd.Flags().Lookup("fail-fast"))
	viper.BindPFlag("cluster", provisionCmd.Flags().Lookup("cluster"))
}
