package stash

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/dabblebox/stash/component/action"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/monitor"
	"github.com/gookit/color"
)

// GetOptions defines the criteria used to download
// configuration files from a service.
type GetOptions struct {
	// Catalog specifies the catalog file name to use.
	// Required: false
	// Default: stash.yml
	Catalog string

	// Files list the file paths to look for in the catalog.
	// Required: false
	// Default: []
	Files []string

	// Tags list the tags to look for in the catalog.
	// Required: false
	// Default: []
	Tags []string

	// Service specifies where to download files from.
	// Required: false
	// Default: any
	Service string

	// Output specifies the format for the file contents.
	// Required: false
	// Default: original
	Output string
}

// Get downloads config files from a remote service.
func Get(opt GetOptions) ([]action.DownloadedFile, error) {
	color.Disable()

	gopt := action.GetOpt{}
	gopt.Files = opt.Files
	gopt.Tags = opt.Tags
	gopt.Service = opt.Service
	gopt.Output = opt.Output

	if len(opt.Catalog) == 0 {
		gopt.Catalog = catalog.DefaultName
	} else {
		gopt.Catalog = opt.Catalog
	}

	m := monitor.New(ioutil.Discard, true)

	files, err := action.Get(gopt, action.Dep{
		Stderr: os.NewFile(0, os.DevNull),
		Stdout: os.NewFile(0, os.DevNull),

		Monitor: &m,
	})

	if len(m.Errors) > 0 {
		return files, m.Errors[0]
	}

	return files, err
}

// GetMap downloads config files from remote services that support maps.
func GetMap(opt GetOptions) (map[string]string, error) {

	downloads, err := Get(opt)
	if err != nil {
		return map[string]string{}, err
	}

	m := map[string]string{}
	for _, d := range downloads {
		tm, err := dotenv.Parse(bytes.NewReader(d.Data))
		if err != nil {
			return m, err
		}

		for k, v := range tm {
			m[k] = v
		}
	}

	return m, nil
}
