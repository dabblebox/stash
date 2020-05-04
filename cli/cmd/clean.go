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

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Deletes local configuration files.",
	Long: `
Users can delete tracked configuration files from local 
folders leaving their environment secure.

Stashed configuration remains untouched.

Example: 

$ stash clean config/dev/.env
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		if err := action.Clean(action.Options{
			Catalog: viper.GetString("file"),
			Service: viper.GetString("service"),
			Tags:    viper.GetStringSlice("tags"),
			Files:   filePaths,
		}, action.Dep{
			Monitor: &m,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
			Stdin:   os.Stdin,
		}); err != nil {
			m.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	cleanCmd.Flags().StringP("service", "s", "", "cloud service")
	cleanCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
}
