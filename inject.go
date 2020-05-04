package stash

import (
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/gookit/color"
)

// InjectOptions defines the criteria used to inject
// remote service configuration into local files.
type InjectOptions struct {
	// Files list the target files.
	// Required: true
	// Default: []
	Files []string

	// Service specifies which service to use.
	// Required: true
	// Default: secrets-manager
	Service string

	// Output specifies the format for the file contents.
	// Required: false
	// Default: original
	Output string
}

// Inject replaces local file tokens with values from a remote service.
func Inject(opt InjectOptions) ([]action.DownloadedFile, error) {
	color.Disable()

	gopt := action.InjectOpt{}
	gopt.Files = opt.Files
	gopt.Service = opt.Service
	gopt.Output = opt.Output

	m := monitor.New(os.Stderr, true)

	files, err := action.Inject(gopt, action.Dep{
		Stderr: os.NewFile(0, os.DevNull),
		Stdout: os.NewFile(0, os.DevNull),

		Monitor: &m,
	})

	if len(m.Errors) > 0 {
		return files, m.Errors[0]
	}

	return files, err
}
