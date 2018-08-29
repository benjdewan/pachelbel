package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/benjdewan/pachelbel/config"
	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/runner"
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

	cxn, err := connection.New(viper.GetString("api-key"),
		viper.GetString("log-file"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := cxn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	cfg, err := readConfigs(cxn, args)
	if err != nil {
		log.Fatal(err)
	} else if len(cfg.Runners) == 0 {
		fmt.Println("Nothing to do")
		return
	}

	if err := runner.
		NewController(cxn, viper.GetBool("dry-run")).
		Run(cfg.Runners); err != nil {
		log.Fatal(err)
	}

	writeOutput(cxn, cfg.EndpointMap)
}

func writeOutput(cxn *connection.Connection, endpointMap map[string]string) {
	dst := viper.GetString("output")
	if err := cxn.ConnectionYAML(endpointMap, dst); err != nil {
		log.Fatal(err)
	}
}

func readConfigs(cxn *connection.Connection, paths []string) (*config.Config, error) {
	var err error
	config.Databases, err = cxn.SupportedDatabases()
	if err != nil {
		return nil, err
	}

	config.Clusters, err = cxn.Clusters()
	if err != nil {
		return nil, err
	}

	config.Datacenters, err = cxn.Datacenters()
	if err != nil {
		return nil, err
	}

	config.CXN = cxn

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
}

func addClusterFlag() {
	provisionCmd.Flags().StringSliceP("cluster", "c", []string{},
		`By default pachelbel provision will provision
				 every deployment provided. Use this flag to
				 limit pachelbel to only process deployments
				 to the specified cluster.

				 This flag can be repeated to specify multiple
				 clusters`)
	if err := viper.BindPFlag("cluster", provisionCmd.Flags().Lookup("cluster")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addDatacenterFlag() {
	provisionCmd.Flags().StringSliceP("datacenter", "d", []string{},
		`By default pachelbel provision will
				 provision every deployment provided. Use this
				 flat to limit pachelbel to only process
				 deployments to the specified datacenter.

				 This flag can be repeated to specify multiple
				 datacenters.`)
	if err := viper.BindPFlag("datacenter", provisionCmd.Flags().Lookup("datacenter")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addOutputFlag() {
	provisionCmd.Flags().StringP("output", "o", "./connection-info.yml",
		`The file to write connection string
				 information to.`)
	if err := viper.BindPFlag("output", provisionCmd.Flags().Lookup("output")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
