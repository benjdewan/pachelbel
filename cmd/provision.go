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
	"os"

	"github.com/benjdewan/pachelbel/config"
	"github.com/benjdewan/pachelbel/connection"
	"github.com/golang-collections/go-datastructures/queue"
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
	Run: runProvision,
}

func runProvision(cmd *cobra.Command, args []string) {
	assertCanStart(args)

	deployments, err := readConfigs(args)
	if err != nil {
		log.Fatal(err)
	}

	cxn, err := connection.Init(viper.GetString("api-key"),
		viper.GetInt("polling-interval"))
	if err != nil {
		log.Fatal(err)
	}

	provision(cxn, deployments)

	writeConnectionStrings(cxn, viper.GetString("output"))
}

func provision(cxn *connection.Connection, deployments []connection.Deployment) {
	errQueue := queue.New(int64(len(deployments)))
	cxn.Provision(deployments, errQueue)
	flush(errQueue)
}

func writeConnectionStrings(cxn *connection.Connection, file string) {
	errQueue := queue.New(0)
	cxn.ConnectionStringsYAML(viper.GetString("output"), errQueue)
	flush(errQueue)
}

func readConfigs(paths []string) ([]connection.Deployment, error) {
	config.BuildClusterFilter(viper.GetStringSlice("cluster"))
	config.BuildDatacenterFilter(viper.GetStringSlice("datacenter"))

	return config.ReadFiles(paths, viper.GetBool("verbose"))
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
	addPollingIntervalFlag()

}

func addClusterFlag() {
	provisionCmd.Flags().StringSliceP("cluster", "c", []string{},
		`By default pachelbel provision will provision every deployment
			provided. Use this flag to limit pachelbel to only
			process deployments to the specified cluster.

			This flag can be repeated to specify multiple clusters`)
	viper.BindPFlag("cluster", provisionCmd.Flags().Lookup("cluster"))
}

func addDatacenterFlag() {
	provisionCmd.Flags().StringSliceP("datacenter", "d", []string{},
		`By default pachelbel provision will provision every
			deployment provided. Use this flat to limit pachelbel
			to only process deployments to the specified
			datacenter.

			This flag can be repeated to specify multiple datacenters.`)
	viper.BindPFlag("datacenter", provisionCmd.Flags().Lookup("datacenter"))
}

func addOutputFlag() {
	provisionCmd.Flags().StringP("output", "o", "./connection-strings.yml",
		`The file to write connection string information to.`)
	viper.BindPFlag("output", provisionCmd.Flags().Lookup("output"))
}

func addPollingIntervalFlag() {
	provisionCmd.Flags().IntP("polling-interval", "p", 5,
		`The polling interval, in seconds, to use when
			waiting for a provisioning recipe to complete`)
	viper.BindPFlag("polling-interval", provisionCmd.Flags().Lookup("polling-interval"))
}

func flush(errQueue *queue.Queue) {
	if errQueue.Empty() {
		errQueue.Dispose()
		return
	}
	items, qErr := errQueue.Get(errQueue.Len())
	if qErr != nil {
		// Get() only returns an error if Dispose() has already
		// been called on the queue.
		panic(qErr)
	}
	for _, unknown := range items {
		switch item := unknown.(type) {
		case error:
			log.Printf("Error: %v", item)
		default:
			log.Fatalf("Only errors should be in the error queue. Found %v", item)
		}
	}
	os.Exit(1)
	errQueue.Dispose()
}
