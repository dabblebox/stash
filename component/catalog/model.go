package catalog

import (
	"fmt"

	"github.com/dabblebox/stash/component/path"
	"github.com/dabblebox/stash/component/service"
)

// DefaultName is the default catalog name.
const DefaultName = "stash.yml"

// Catalog is reference file for the stashed configuration files.
type Catalog struct {
	Version string `yaml:"version"`

	Context string `yaml:"context"`

	AutoClean bool `yaml:"clean" mapstructure:"clean"`

	Files map[string]File `yaml:"files"`
}

// GetFile ...
func (c *Catalog) GetFile(path string) (File, bool) {

	for _, f := range c.Files {
		if path == f.Path {
			return f, true
		}
	}

	return File{}, false
}

// AddFile ...
func (c *Catalog) AddFile(key, filePath, service string, tags []string) error {
	if len(key) == 0 {
		key = formatInitialKey(filePath)
	}

	if _, found := c.Files[key]; found {
		return fmt.Errorf("catalog key %s already exists", key)
	}

	if len(tags) == 0 {
		tags = path.Tags(filePath)
	}

	f := File{
		Path:    filePath,
		Service: service,
		Type:    path.Type(filePath),
		Tags:    tags,
	}

	if c.Files == nil {
		c.Files = map[string]File{}
	}

	c.Files[key] = f

	return nil
}

// Filter ...
func (c *Catalog) Filter(filter Filter) map[string]File {
	filtered := map[string]File{}

	for k, f := range c.Files {

		if f.Matches(filter) {
			filtered[k] = f
		}
	}

	return filtered
}

// MergeResults ...
func (c *Catalog) MergeResults(files []service.File) {

	for _, sf := range files {
		if f, ok := c.Files[sf.CatalogKey]; ok {
			f.Keys = sf.Keys
			f.Options = sf.Options

			c.Files[sf.CatalogKey] = f
		}
	}
}

func (c *Catalog) LookupService(filePath string) string {
	t := path.Type(filePath)

	for _, f := range c.Files {
		if f.Type == t {
			return f.Service
		}
	}

	return ""
}
