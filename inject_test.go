package stash

import (
	"fmt"
	"os"
	"testing"

	"github.com/dabblebox/stash/component/output"
)

func TestInject(t *testing.T) {

	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_PROFILE", "devops")

	opt := InjectOptions{}
	opt.Output = output.TypeOriginal
	opt.Files = []string{"sample/dev/.env"}
	opt.Service = "secrets-manager"

	files, err := Inject(opt)
	if err != nil {
		t.Error(err)
	}

	for _, f := range files {
		fmt.Printf("%s\n", string(f.Data))
	}
}
