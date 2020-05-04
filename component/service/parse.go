package service

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dabblebox/stash/component/dotenv"
)

func (f *File) parseJSON() (map[string]string, error) {
	m := map[string]string{}

	tempMap := map[string]json.RawMessage{}

	if err := json.Unmarshal(f.Data, &tempMap); err != nil {
		return m, err
	}

	for k, v := range tempMap {
		t := make(map[string]string)

		if err := json.Unmarshal(v, &t); err != nil {
			m[k] = fmt.Sprintf(`{"%s":%+v}`, k, string(v))
		} else {
			m[k] = fmt.Sprintf(`%+v`, string(v))
		}
	}

	return m, nil
}

func (f *File) parseENV(expandVariables bool) (map[string]string, error) {
	return dotenv.Parse(bytes.NewReader(f.Data), expandVariables)
}
