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
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/editor"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/dabblebox/stash/component/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration.",
	Long: `
Users can edit configuration files from a cloud service. 
Once the edits are complete and the file is closed, the 
changes are synced with the cloud service.

Example: 

$ stash edit config/dev/.env
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		dep := action.Dep{
			Monitor: &m,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
			Stdin:   os.Stdin,
		}

		o := action.GetOpt{}

		o.Files = filePaths
		o.Output = output.TypeFile
		o.Catalog = viper.GetString("file")
		o.Service = viper.GetString("service")
		o.Tags = viper.GetStringSlice("tags")

		filter := catalog.NewGetFilter(o.Files, o.Tags, o.Service)

		if filter.Empty() {
			files, err := action.Browse(o.Catalog, dep)
			if err != nil {
				m.Fatal(err)
			}

			if len(files) == 0 {
				os.Exit(0)
			}

			o.Files = files
		}

		downloaded, err := action.Get(o, dep)

		if err != nil {
			m.Fatal(err)
		}

		edited := []string{}
		for _, df := range downloaded {

			if err := file.Write(df.Path, df.Data); err != nil {
				m.Fatal(err)
			}

			if err := editor.Run(df.Path); err != nil {
				m.Fatal(err)
			}

			edited = append(edited, df.Path)
		}

		opts := action.SyncOpt{}
		opts.Catalog = viper.GetString("file")
		opts.Files = edited

		if err := action.Sync(opts, dep); err != nil {
			m.Fatal(err)
		}

		if len(edited) > 0 {
			fmt.Fprintf(os.Stderr, "Remember to modify infrastructure when applicable.\n\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	editCmd.Flags().StringP("service", "s", "", "cloud service")
	editCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")
}
