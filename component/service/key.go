package service

import (
	"fmt"
	"strings"
)

// FormatObjectKey ...
func FormatObjectKey(context, path string, service IService) string {
	return service.ObjectKey(fmt.Sprintf("%s/%s", context, strings.TrimLeft(path, "./")))
}
