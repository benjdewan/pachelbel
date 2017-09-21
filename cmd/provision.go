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

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Idempotent provisioner of compose deployments",
	Long: `pachelbel provision reads in YAML configuration(s) describing a list of
deployments that should exist in a list of clusters of specified sizes,
and ensures they do.

If the deployments do not exist, they are created. If they exist, but are
the wrong size they are scaled. If they are deployed as specified in the
configuration no actions are taken.`,
	Run: runProvision,
}

func runProvision(cmd *cobra.Command, args []string) {
	assertCanStart(args)

	cfg, err := readConfigs(args)
	if err != nil {
		log.Fatal(err)
	}

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

	process(cxn, cfg.Accessors)

	writeOutput(cxn, cfg.EndpointMap)
}

func process(cxn *connection.Connection, accessors []connection.Accessor) {
	if err := cxn.Process(accessors); err != nil {
		log.Fatal(err)
	}
}

func writeOutput(cxn *connection.Connection, endpointMap map[string]string) {
	dst := viper.GetString("output")
	if err := cxn.ConnectionYAML(endpointMap, dst); err != nil {
		log.Fatal(err)
	}
}

func readConfigs(paths []string) (*config.Config, error) {
	config.BuildClusterFilter(viper.GetStringSlice("cluster"))
	config.BuildDatacenterFilter(viper.GetStringSlice("datacenter"))

	return config.ReadFiles(paths)
}

func assertCanStart(args []string) {
	if len(args) == 0 {
		log.Fatal("The 'provision' command requires at least one configuration file or directory as input")
	}
}

func init() {
	RootCmd.AddCommand(provisionCmd)
	addClusterFlag()
	addDatacenterFlag()
	addOutputFlag()
	addDryRunFlag()
	addLogFileFlag()
}

func addClusterFlag() {
	provisionCmd.Flags().StringSliceP("cluster", "c", []string{},
		`By default pachelbel provision will provision
				 every deployment provided. Use this flag to
				 limit pachelbel to only process deployments
				 to the specified cluster.

				 This flag can be repeated to specify multiple
				 clusters`)
	viper.BindPFlag("cluster", provisionCmd.Flags().Lookup("cluster"))
}

func addDatacenterFlag() {
	provisionCmd.Flags().StringSliceP("datacenter", "d", []string{},
		`By default pachelbel provision will
				 provision every deployment provided. Use this
				 flat to limit pachelbel to only process
				 deployments to the specified datacenter.

				 This flag can be repeated to specify multiple
				 datacenters.`)
	viper.BindPFlag("datacenter", provisionCmd.Flags().Lookup("datacenter"))
}

func addOutputFlag() {
	provisionCmd.Flags().StringP("output", "o", "./connection-info.yml",
		`The file to write connection string
				 information to.`)
	viper.BindPFlag("output", provisionCmd.Flags().Lookup("output"))
}

func addDryRunFlag() {
	provisionCmd.Flags().BoolP("dry-run", "n", false,
		`Simulate a provision run without making any
				 real changes.

				 If a deployment already exists the connection
				 strings to it will be returned, but any
				 additional steps to rescale, upgrade or add
				 team roles to the deployment will be ignored.

				 If a deployment does not exist it will not be
				 created and a fake connection string will be
				 returned for testing purposes.`)
	viper.BindPFlag("dry-run", provisionCmd.Flags().Lookup("dry-run"))
}

func addLogFileFlag() {
	provisionCmd.Flags().StringP("log-file", "l", "",
		`If specified pachelbel will enable logging for
				all Compose API requests and write them, as well
				as the reponses, to the specified log file`)
	viper.BindPFlag("log-file", provisionCmd.Flags().Lookup("log-file"))
}
