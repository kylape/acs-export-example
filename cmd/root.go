package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kylape/acs-export-example/pkg/export"
	"github.com/kylape/acs-export-example/pkg/table"
)

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

		err = table.RenderTable(deployments, images)
		if err != nil {
			panic(errors.Wrap(err, "Failed to render table"))
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
