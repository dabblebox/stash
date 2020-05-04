package catalog

import (
	"encoding/json"
	"os"
)

func ExpandEnv(c Catalog) (Catalog, error) {

	d, err := json.Marshal(c)
	if err != nil {
		return c, err
	}

	d = []byte(os.ExpandEnv(string(d)))

	if err := json.Unmarshal(d, &c); err != nil {
		return c, err
	}

	return c, nil
}
