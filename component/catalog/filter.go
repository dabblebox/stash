package catalog

import (
	"fmt"
	"strings"
)

// Filter ...
type Filter struct {
	Files []string
	Tags  []string

	Service string
}

// Empty ...
func (f Filter) Empty() bool {
	return len(f.Files) == 0 && len(f.Tags) == 0 && len(f.Service) == 0
}

// Format ...
func (f Filter) Format(delimiter string) string {
	var b strings.Builder

	if len(f.Files) > 0 {
		b.WriteString(fmt.Sprintf("file(s)%v%s", f.Files, delimiter))
	}

	if len(f.Tags) > 0 {
		b.WriteString(fmt.Sprintf("tag(s)%v%s", f.Tags, delimiter))
	}

	if len(f.Service) > 0 {
		b.WriteString(fmt.Sprintf("service[%s]%s", f.Service, delimiter))
	}

	return strings.Trim(b.String(), delimiter)
}

// NewPushFilter ...
func NewPushFilter(files, tags []string, service string) Filter {
	filter := Filter{Files: files}

	if filter.Empty() {
		filter.Tags = tags
		filter.Service = service
	}

	return filter
}

// NewGetFilter ...
func NewGetFilter(files, tags []string, service string) Filter {
	filter := Filter{
		Files:   files,
		Tags:    tags,
		Service: service,
	}

	return filter
}
