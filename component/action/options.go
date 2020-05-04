package action

import (
	"os"

	"github.com/dabblebox/stash/component/monitor"
)

// Options ...
type Options struct {
	Catalog string

	Files []string
	Tags  []string

	Service string

	Warn bool
}

// Dep ...
type Dep struct {
	Stderr *os.File
	Stdout *os.File
	Stdin  *os.File

	Monitor *monitor.Monitor
}
