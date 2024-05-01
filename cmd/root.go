package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kylape/acs-export-example/pkg/config"
	"github.com/kylape/acs-export-example/pkg/csv"
	"github.com/kylape/acs-export-example/pkg/export"
	"github.com/kylape/acs-export-example/pkg/filter"
	"github.com/kylape/acs-export-example/pkg/table"
)

var cfg = config.ConfigType{}

var rootCmd = &cobra.Command{
	Use:   "acs-export-example",
	Short: "Use the ACS export APIs",
	Long:  `CLI to browse data pulled from ACS (Advanced Cluster Security) (i.e. StackRox).`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		exporter, err := export.New(ctx)
		if err != nil {
			panic(errors.Wrap(err, "could not create exporter"))
		}

		println("Fetching deployments")
		deployments, err := exporter.GetDeployments(cfg)
		if err != nil {
			panic(errors.Wrap(err, "could not get deployments"))
		}

		println("Fetching images")
		images, err := exporter.GetImages(cfg)
		if err != nil {
			panic(errors.Wrap(err, "could not get images"))
		}

		deployments, images, err = filter.Filter(deployments, images, cfg)

		if cfg.Output == "table" {
			if err = table.RenderTable(deployments, images); err != nil {
				panic(errors.Wrap(err, "Failed to render table"))
			}
		} else if cfg.Output == "csv" {
			if err = csv.RenderCsv(deployments, images); err != nil {
				panic(errors.Wrap(err, "Failed to render table"))
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfg.Output, "output", "o", "table", "Output format.  Available options: [table, csv]")
	rootCmd.PersistentFlags().StringVarP(&cfg.NamespaceFilter, "namespace", "n", "", "Namespace filter. Filtered client-side.")
	rootCmd.PersistentFlags().StringVarP(&cfg.ClusterFilter, "cluster", "c", "", "Cluster filter. Filtered client-side.")
	rootCmd.PersistentFlags().StringVarP(&cfg.ImageNameFilter, "image", "i", "", "Image name filter. Filtered client-side.")
	rootCmd.PersistentFlags().StringVarP(&cfg.VulnerabilityFilter, "vuln", "v", "", "Vulnerability filter. Filtered client-side.")
	rootCmd.PersistentFlags().StringVarP(&cfg.QueryFilter, "query", "q", "", "Pass a query string to the server")
}
