package search

import (
	"os"
	"path/filepath"
	"regexp"
)

func By(regex string) ([]string, error) {
	results := []string{}

	re, err := regexp.Compile(regex)
	if err != nil {
		return results, err
	}

	if err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if err != nil {
				return err
			}
			if re.MatchString(path) {
				results = append(results, path)
			}
			return nil
		}); err != nil {
		return results, err
	}

	return results, nil
}
