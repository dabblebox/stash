package path

import (
	"path/filepath"
	"strings"
)

// Type ...
func Type(path string) string {
	ext := filepath.Ext(path)

	replacer := strings.NewReplacer(".", "")

	return replacer.Replace(ext)
}
