package path

import (
	"path/filepath"
	"strings"
)

// Tags ...
func Tags(path string) []string {
	ext := filepath.Ext(path)

	path = strings.TrimSuffix(path, ext)

	path = strings.Trim(path, "./")

	path = strings.ReplaceAll(path, "../", "")

	tags := strings.Split(path, "/")

	if len(tags) <= 1 {
		return []string{}
	}

	return tags[0:]
}
