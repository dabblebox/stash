/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists cataloged files.",
	Long: `
Users can list configuration files tracked by
the catalog, "stash.yml".

This command does not verify the cloud service
is in sync with the local files.

Example: 

$ stash list -t dev
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		if err := action.List(action.Options{
			Catalog: viper.GetString("file"),
			Service: viper.GetString("service"),
			Tags:    viper.GetStringSlice("tags"),
			Files:   filePaths,
		}, action.Dep{
			Monitor: &m,
			Stdin:   os.Stdin,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
		}); err != nil {
			m.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	listCmd.Flags().StringP("service", "s", "", "cloud service")
	listCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
}
