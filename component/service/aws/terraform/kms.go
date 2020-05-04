package terraform

import (
	"os"
	"text/template"

	"github.com/dabblebox/stash/component/service/aws/kms"
)

func EnsureTFKMSFile(filePath, kmsKeyID string) error {

	if _, err := os.Stat(filePath); err != nil {
		return CreateTFKMSFile(filePath, kmsKeyID)
	}

	return nil
}

func CreateTFKMSFile(filePath, kmsKeyID string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New("kms").Parse(kms.HCLTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, kms.HCLModel{
		KMSKeyID: kmsKeyID,
	}); err != nil {
		return err
	}

	return nil
}
