package terraform

import (
	"os"

	"github.com/dabblebox/stash/component/service/aws/vars"
)

func EnsureTFVariablesFile(filePath string) error {

	if _, err := os.Stat(filePath); err != nil {

		return CreateTFVariablesFile(filePath)
	}

	return nil
}

func CreateTFVariablesFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(vars.HCLVariablesTemplate)

	return err
}
