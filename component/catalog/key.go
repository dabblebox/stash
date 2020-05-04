package catalog

import (
	"regexp"
	"strings"
)

// formatInitialKey is used for a default when a new file
// is being cataloged. This allows the key to be customized
// later by a user after creation.
//
// This is also used to generate the key for the user state
// file that assists in change detection.
//
// This function should NOT be used when comparing values to
// the catalog file key since it can be changed by a user.
func formatInitialKey(filePath string) string {

	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)

	return strings.ToLower(strings.Trim(re.ReplaceAllString(filePath, "_"), "_"))
}
