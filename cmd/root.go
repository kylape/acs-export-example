package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kylape/acs-export-example/pkg/csv"
	"github.com/kylape/acs-export-example/pkg/export"
	"github.com/kylape/acs-export-example/pkg/table"
)

type configType struct {
	output string
}

var config = configType{}

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
		deployments, err := exporter.GetDeployments()
		if err != nil {
			panic(errors.Wrap(err, "could not get deployments"))
		}

		println("Fetching images")
		images, err := exporter.GetImages()
		if err != nil {
			panic(errors.Wrap(err, "could not get images"))
		}

		if config.output == "table" {
			if err = table.RenderTable(deployments, images); err != nil {
				panic(errors.Wrap(err, "Failed to render table"))
			}
		} else if config.output == "csv" {
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
	rootCmd.PersistentFlags().StringVarP(&config.output, "output", "o", "table", "Output format.  Available options: [table, csv]")
}
