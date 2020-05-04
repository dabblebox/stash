package catalog

import (
	"os"

	"github.com/dabblebox/stash/component/service"
	"github.com/dabblebox/stash/component/slice"
)

// File represents a configuration file that is remotely stashed. All fields
// in this model should be safe to be shared and checked into source control.
type File struct {
	// Path is the local location.
	Path string `yaml:"path"`

	// Type is the local file type. (e.g. env,json,txt,etc.)
	Type string `yaml:"type,omitempty"`

	// Service specifies the remote service.
	Service string `yaml:"service,omitempty"`

	// Options allow services to persist user preferences locally.
	Options map[string]string `yaml:"opt,omitempty" mapstructure:"opt"`

	// Keys allows services to track which fields are stashed.
	Keys []string `yaml:"keys,omitempty"`

	// Tags allow files to be grouped; so, they can be listed, purged,
	// and restored in a single command.
	Tags []string `yaml:"tags,omitempty"`

	// Clean deletes the local files after changes have been pushed
	// to the remote service.
	Clean bool `yaml:"clean,omitempty"`
}

// RemoveTag ...
func (f *File) RemoveTag(tag string) {
	idx := -1

	for i, k := range f.Tags {
		if k == tag {
			idx = i
		}
	}

	if idx != -1 {
		f.Tags = slice.Remove(f.Tags, idx)
	}
}

// AddTag ...
func (f *File) AddTag(tag string) {
	for _, k := range f.Tags {
		if k == tag {
			return
		}
	}

	f.Tags = append(f.Tags, tag)
}

// ToServiceModel ...
func (f File) ToServiceModel(context, key string, remote service.IService, data []byte) (service.File, error) {

	state, err := f.LookupState(context)
	if err != nil {
		if !os.IsNotExist(err) {
			return service.File{}, err
		}
	}

	if f.Options == nil {
		f.Options = map[string]string{}
	}

	return service.File{
		Context:    context,
		CatalogKey: key,
		LocalPath:  f.Path,
		RemoteKey:  service.FormatObjectKey(context, f.Path, remote),
		Type:       f.Type,
		Keys:       f.Keys,
		Options:    f.Options,
		Data:       data,
		Synced:     state.Synced,
	}, nil
}

// Matches ...
func (f *File) Matches(filter Filter) bool {

	if len(filter.Files) > 0 {
		if !slice.AnyIn([]string{f.Path}, filter.Files) {
			return false
		}
	}

	if len(filter.Tags) > 0 {
		if !slice.Subset(filter.Tags, f.Tags) {
			return false
		}
	}

	if len(filter.Service) > 0 {
		if filter.Service != f.Service {
			return false
		}
	}

	return true
}

// GroupByService ...
func GroupByService(files map[string]File) map[string]map[string]File {
	grouped := map[string]map[string]File{}

	for k, f := range files {
		sg, ok := grouped[f.Service]
		if ok {
			sg[k] = f
			grouped[f.Service] = sg
		} else {
			grouped[f.Service] = map[string]File{
				k: f,
			}
		}
	}

	return grouped
}

// GroupByServiceSlice ...
func GroupByServiceSlice(files map[string]File) map[string][]File {
	grouped := map[string][]File{}

	for _, f := range files {
		sg, ok := grouped[f.Service]
		if ok {
			sg = append(sg, f)
			grouped[f.Service] = sg
		} else {
			grouped[f.Service] = []File{f}
		}
	}

	return grouped
}
