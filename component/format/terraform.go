package format

import (
	"regexp"
	"strings"
)

func TerraformResourceName(value string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)

	return strings.ToLower(strings.Trim(re.ReplaceAllString(value, "_"), "_"))
}
