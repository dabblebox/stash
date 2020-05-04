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
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/dabblebox/stash/component/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// injectCmd represents the inject command
var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Replaces keys in a file from a cloud service.",
	Long: `
Users can replace tokens in a file with values from a cloud
service when the stashed data is stored in a cloud service
that supports key/value pairs.

Example: 

$ stash inject config/dev/.env
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, filePaths []string) {
		m := monitor.New(os.Stderr, viper.GetBool("logs"))

		downloaded, err := action.Inject(action.InjectOpt{
			Files:   filePaths,
			Service: viper.GetString("service"),
			Output:  viper.GetString("output"),
		}, action.Dep{
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
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringP("output", "o", "", "file output format")
	injectCmd.Flags().StringP("service", "s", "", "cloud service")

	injectCmd.MarkFlagRequired("service")

	viper.SetDefault("output", output.TypeOriginal)
}
