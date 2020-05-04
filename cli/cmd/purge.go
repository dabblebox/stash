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
	"fmt"
	"log"
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Deletes configuration from a cloud service.",
	Long: `
Users can delete configuration files from a cloud service. 
Each purge deletes cloud service configuration leaving local 
configuration files untouched when present.

Example: 

$ stash purge config/dev/.env
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		remaining, _, err := action.Purge(action.Options{
			Catalog: viper.GetString("file"),
			Service: viper.GetString("service"),
			Tags:    viper.GetStringSlice("tags"),
			Files:   filePaths,
			Warn:    viper.GetBool("warn"),
		}, action.Dep{
			Monitor: &m,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
			Stdin:   os.Stdin,
		})

		if err != nil {
			m.Fatal(err)
		}

		if remaining == 0 {
			if err := os.Remove(viper.GetString("file")); err != nil {
				log.Fatal(err)
			}

			fmt.Fprintf(os.Stderr, "Remember to clean up any configuration infrastructure.\n\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	purgeCmd.Flags().StringP("service", "s", "", "cloud service")
	purgeCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
	purgeCmd.Flags().BoolP("warn", "w", true, "disable warnings")
}
