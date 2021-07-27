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
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/dabblebox/stash/component/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Downloads configuration from a cloud service.",
	Long: `
Users can download configuration files from a cloud service. 
Downloads can either restore the file to the original folder 
location or send the data to stdout after applying an optional
data transformation.

Examples: 
  stash get -t dev
  stash get config/dev/.env -o file 

Outputs:
  file                  	file    system	original file
  terraform             	file    system	terraform scripts
  ecs-task-inject-json  	stdout  AWS ECS task definition secrets / envfile (JSON) (key/arn)
  ecs-task-inject-env   	stdout  AWS ECS task definition secrets / envfile (ENV) (key/arn)
  ecs-task-env          	stdout  AWS ECS task definition environment (JSON) (key/value)
  json                  	stdout  JSON object
  terminal-export       	stdout  prepend "export " to each key/value pair (double quotes)
  terminal-export-literal   stdout  prepend "export " to each key/value pair (single quotes)
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		opts := action.GetOpt{}

		opts.Files = filePaths
		opts.Catalog = viper.GetString("file")
		opts.Service = viper.GetString("service")
		opts.Tags = viper.GetStringSlice("tags")
		opts.Output = viper.GetString("output")
		
		downloaded, err := action.Get(opts, action.Dep{
			Monitor: &m,
			Stderr:  os.Stderr,
			Stdout:  os.Stdout,
			Stdin:   os.Stdin,
		})

		if err != nil {
			m.Fatal(err)
		}

		pipe := bytes.Buffer{}
		for _, df := range downloaded {

			switch df.Output {
			case output.TypeFile:
				if err := file.Write(df.Path, df.Data); err != nil {
					m.Fatal(err)
				}
			case output.TypeTerraform:
				tfFilePath := fmt.Sprintf("%s/%s-%s.tf", filepath.Dir(df.Path), df.Service, file.DashIt(filepath.Base(df.Path)))
				if err := file.Write(tfFilePath, df.Data); err != nil {
					m.Fatal(err)
				}
			default:
				df.Data = append(df.Data, []byte("\n")...)

				if _, err := pipe.Write(df.Data); err != nil {
					m.Fatal(err)
				}
			}
		}

		if _, err := pipe.WriteTo(os.Stdout); err != nil {
			m.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("file", "f", catalog.DefaultName, "catalog name")
	getCmd.Flags().StringP("output", "o", "original", "output format")
	getCmd.Flags().StringP("service", "s", "", "cloud service")
	getCmd.Flags().StringSliceP("tags", "t", []string{}, "tagging for quick file reference")

	viper.SetDefault("output", output.TypeOriginal)
}
