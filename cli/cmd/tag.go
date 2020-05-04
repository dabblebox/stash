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

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Tags cataloged files.",
	Long: `
Users can tag configuration files tracked by the catalog, 
"stash.yml" for performaing group actions.

This command only alters the local catalog, "stash.yml".

Example: 

$ stash tag config/dev/.env -t app
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		opts := action.TagOpt{}

		opts.Files = filePaths
		opts.Catalog = viper.GetString("file")
		opts.Service = viper.GetString("service")
		opts.Tags = viper.GetStringSlice("tags")
		opts.Add = viper.GetStringSlice("add")
		opts.Delete = viper.GetStringSlice("delete")

		if err := action.Tag(opts, action.Dep{
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
	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	tagCmd.Flags().StringP("service", "s", "", "cloud service")
	tagCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
	tagCmd.Flags().StringSliceP("add", "a", []string{}, "remove tag")
	tagCmd.Flags().StringSliceP("delete", "d", []string{}, "delete tag")
}
