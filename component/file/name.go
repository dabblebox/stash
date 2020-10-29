package file

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Number(file string) string {

	num := 0

	ext := filepath.Ext(file)
	original := strings.Replace(filepath.Base(file), ext, "", -1)
	dir := filepath.Dir(file)

	name := original
	for {
		if _, err := os.Stat(fmt.Sprintf("%s/%s%s", dir, name, ext)); err != nil {
			break
		}

		num++

		name = fmt.Sprintf("%s-%d", original, num)
	}

	return fmt.Sprintf("%s/%s%s", dir, name, ext)
}

func DashIt(path string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)

	return re.ReplaceAllString(path, "-")
}
