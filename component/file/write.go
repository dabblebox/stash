package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Write ...
func Write(path string, b []byte) error {
	if strings.Contains(path, "/") {
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path, b, os.ModePerm)
}
