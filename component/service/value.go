package service

import (
	"regexp"
	"strings"
)

func isString(value string) bool {
	number := regexp.MustCompile(`^[0-9.]+$`)

	if number.MatchString(value) {
		return false
	}

	boolean := regexp.MustCompile(`^true$|^false$`)

	if boolean.MatchString(strings.ToLower(value)) {
		return false
	}

	return true
}
