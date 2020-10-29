/*
Copyright © 2020 NAME HERE <EMAIL syncRESS>

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
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs local configuration files to a cloud service.",
	Long: `
Users can sync configuration files with a cloud service creating a 
local catalog, "stash.yml". The catalog can be safely shared or 
checked in to source control allowing other users or applications 
to access the stash.

Local configuration files should *NEVER* be checked into source 
control when they contain secrets.

Example: 

$ stash sync config/dev/.env
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		opts := action.SyncOpt{}

		opts.Files = filePaths
		opts.Catalog = viper.GetString("file")
		opts.Service = viper.GetString("service")
		opts.Tags = viper.GetStringSlice("tags")
		opts.Context = viper.GetString("context")

		if err := action.Sync(opts,
			action.Dep{
				Monitor: &m,
				Stderr:  os.Stderr,
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
			}); err != nil {
			m.Fatal(err)
		}

		fmt.Fprintf(os.Stderr, "Remember to modify infrastructure when applicable.\n\n")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	syncCmd.Flags().StringP("context", "c", "", "cloud storage key prefix")
	syncCmd.Flags().StringP("service", "s", "", "cloud service")
	syncCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
}
