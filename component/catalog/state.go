package catalog

import (
	"fmt"
	"os"
	"time"

	"github.com/dabblebox/stash/component/file"
	"gopkg.in/yaml.v2"
)

const fileName = "state.yml"

// RecordState ...
func (f *File) RecordState(context string) error {
	files := map[string]State{}

	b, err := file.Read(file.HomePath(fileName))
	if err == nil {
		if err = yaml.Unmarshal(b, &files); err != nil {
			return err
		}
	}

	// Add seconds to account for remote service timestamp delays.
	files[formatStateKey(context, f.Path)] = State{
		Synced: time.Now().UTC().Add(10 * time.Second),
	}

	b, err = yaml.Marshal(files)
	if err != nil {
		return err
	}

	return file.Write(file.HomePath(fileName), b)
}

// LookupState ...
func (f *File) LookupState(context string) (State, error) {
	files := map[string]State{}

	b, err := file.Read(file.HomePath(fileName))
	if err != nil {
		return State{}, err
	}

	if err = yaml.Unmarshal(b, &files); err != nil {
		return State{}, err
	}

	if s, ok := files[formatStateKey(context, f.Path)]; ok {
		return s, nil
	}

	return State{}, os.ErrNotExist
}

func formatStateKey(context, path string) string {
	return fmt.Sprintf("%s_%s", context, formatInitialKey(path))
}

type State struct {
	Synced time.Time
}
