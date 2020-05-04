package file

import (
	"fmt"
	"log"

	"github.com/mitchellh/go-homedir"
)

func HomePath(path string) string {
	const appFolder = ".stash"

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s/%s/%s", home, appFolder, path)
}
