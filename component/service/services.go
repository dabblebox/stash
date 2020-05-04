package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
)

const (
	SecurityRatingLow    = 3
	SecurityRatingMedium = 2
	SecurityRatingHigh   = 1
)

var (
	fileTokenColor = color.FgCyan.Render
)

// Services ...
var Services = map[string]IService{}

// ListCompatible returns the remote servcies that can
// hold the specified files.
func ListCompatible(files []string) []IService {

	types := []string{}
	for _, f := range files {
		types = append(types, strings.TrimLeft(filepath.Ext(f), "."))
	}

	compatible := []IService{}
	for _, s := range Services {
		if s.Compatible(types) {
			compatible = append(compatible, s)
		}
	}

	return compatible
}

// ToKeys converts a list of services into a list of services keys.
func ToKeys(servcies []IService) []string {
	keys := []string{}

	for _, s := range servcies {
		keys = append(keys, s.Key())
	}

	return keys
}

func printFile(path string, io *os.File) {
	fmt.Fprintf(io, "  - [%s]\n", fileTokenColor(path))
}

type BySecurityRating []IService

func (s BySecurityRating) Len() int           { return len(s) }
func (s BySecurityRating) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s BySecurityRating) Less(i, j int) bool { return s[i].SecurityRating() < s[j].SecurityRating() }
