package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/slice"
)

// File ...
type File struct {
	// Context
	Context string

	// CatalogKey
	CatalogKey string

	// RemoteKey is the remote service key.
	RemoteKey string

	// LocalPath is the path to the local file.
	LocalPath string

	// Type describes the file type
	Type string

	// Options allow servcies to persist user preferences locally.
	Options map[string]string

	// Keys tracks which fields are stashed.
	Keys []string

	// Data contains the file contents.
	Data []byte

	// Synced is the last time the file was synced with the service.
	Synced time.Time
}

func toEnvVarKey(key string) string {
	return fmt.Sprintf("STASH_%s", strings.ToUpper(key))
}

func (f *File) SupportsParsing() bool {
	switch f.Type {
	case file.TypeEnv:
		return true
	case file.TypeJSON:
		jsArray := regexp.MustCompile(`(?s)^\[.*\]`)

		if jsArray.Match(f.Data) {
			return false
		}

		return true
	}

	return false
}

type Dep struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

type Opt struct {
	Key          string
	DefaultValue string
	Description  string
	Items        []string
}

func (f *File) EnsureOption(opt Opt, io IO) error {
	// Check file options
	if _, ok := f.Options[opt.Key]; ok {
		return nil
	}

	// Check environment variable
	if value, ok := os.LookupEnv(toEnvVarKey(opt.Key)); ok {
		f.Options[opt.Key] = strings.ToLower(value)
		return nil
	}

	// Get from user
	var prompt survey.Prompt

	value := ""
	if opt.Items != nil {
		prompt = &survey.Select{
			Message: opt.Key,
			Options: opt.Items,
		}
	} else {
		prompt = &survey.Input{
			Message: opt.Key,
		}
	}

	if len(opt.DefaultValue) > 0 {
		switch v := prompt.(type) {
		case *survey.Input:
			v.Default = opt.DefaultValue
		case *survey.Select:
			v.Default = opt.DefaultValue
		}
	}

	if len(opt.Description) > 0 {
		switch v := prompt.(type) {
		case *survey.Input:
			v.Help = opt.Description
		case *survey.Select:
			v.Help = opt.Description
		}
	}

	if err := survey.AskOne(prompt, &value,
		survey.WithStdio(io.Stdin, io.Stdout, io.Stderr)); err != nil {
		return err
	}

	f.Options[opt.Key] = value

	return nil
}

// RemoveKey ...
func (f *File) RemoveKey(key string) {
	idx := -1

	for i, k := range f.Keys {
		if k == key {
			idx = i
		}
	}

	if idx != -1 {
		f.Keys = slice.Remove(f.Keys, idx)
	}
}

// AddKey ...
func (f *File) AddKey(key string) {
	for _, k := range f.Keys {
		if k == key {
			return
		}
	}

	f.Keys = append(f.Keys, key)
}

// LookupRemoteKey ...
func (f *File) LookupRemoteKey(prop string) (string, bool) {
	for _, remoteKey := range f.Keys {
		remoteProp := filepath.Base(remoteKey)

		if prop == remoteProp {
			return remoteKey, true
		}
	}

	return "", false
}
