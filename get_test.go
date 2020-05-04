package stash

import (
	"fmt"
	"os"
	"testing"

	"github.com/dabblebox/stash/component/output"
)

func TestGet(t *testing.T) {

	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_PROFILE", "devops")

	opt := GetOptions{}
	opt.Output = output.TypeOriginal
	opt.Tags = []string{"dev"}

	m, err := GetMap(opt)
	if err != nil {
		t.Error(err)
	}

	for k, v := range m {
		fmt.Printf("%s=%s\n", k, v)
	}
}
