package catalog

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Save ...
func Save(file string, c Catalog) error {

	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	comment := `## Stash Catalog ##
`

	d = append([]byte(comment), d...)

	return ioutil.WriteFile(file, d, 0644)

}
