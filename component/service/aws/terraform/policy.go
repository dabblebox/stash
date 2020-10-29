package terraform

import (
	"os"
	"text/template"

	"github.com/dabblebox/stash/component/service/aws/sm"
)

func EnsureTFSMPolicyFile(filePath string) error {

	if _, err := os.Stat(filePath); err != nil {
		return CreateTFSMPolicyFile(filePath)
	}

	return nil
}

func CreateTFSMPolicyFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New("sm").Parse(sm.HCLPolicyTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, nil); err != nil {
		return err
	}

	return nil
}
