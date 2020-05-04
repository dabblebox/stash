package service

import (
	"os"
	"strings"
)

// IService defines the methods a remote service must satisfy.
type IService interface {
	// Key uniquely identifies the service and can be entered
	// as a flag value by CLI users.
	// i.e. $ stash sync .env -s s3
	Key() string

	// ObjectKey uniquely identifies each object stashed.
	// This is called to display each files remote key
	// in the user output.
	ObjectKey(path string) string

	// Compatible determines if the list of file types are
	// supported by the service. File types are the file
	// extensions without the leading dot.
	// i.e. env,json,yml
	Compatible(types []string) bool

	// SecurityRating is a number between one and three where
	// one inicates the most secure.
	SecurityRating() int

	// PreHook is called before Sync, Download, or Purge to
	// initilize the service.
	PreHook(io IO) error

	// Sync uploads file contents to the remote service.
	Sync(file File) (File, error)

	// Download gets file contents from the remote service.
	Download(file File, format string) (File, error)

	// Purge deletes file contents from the remote service.
	Purge(file File) error
}

// IO ...
type IO struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

// Name ...
func Name(key string) string {
	return strings.Title(strings.Replace(key, "-", " ", -1))
}
