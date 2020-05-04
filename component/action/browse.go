package action

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dabblebox/stash/component/catalog"
)

// Browse ...
func Browse(catalogFile string, dep Dep) ([]string, error) {

	c, err := catalog.Read(catalogFile)
	if err != nil {
		return []string{}, fmt.Errorf("%s: %s", catalogFile, err)
	}

	if len(c.Files) == 1 {
		for _, f := range c.Files {
			return []string{f.Path}, nil
		}
	}

	options := []string{}
	for _, f := range c.Files {
		options = append(options, f.Path)
	}

	prompt := &survey.MultiSelect{
		Message:  "Edit",
		Options:  options,
		PageSize: 20,
	}

	files := []string{}
	if err := survey.AskOne(prompt, &files,
		survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr),
	); err != nil {
		return []string{}, err
	}

	return files, nil
}
