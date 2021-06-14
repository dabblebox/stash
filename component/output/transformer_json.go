package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/file"
)

type JSONTransformer struct {
	fileType string
}

func (t JSONTransformer) Transform(data []byte) ([]byte, error) {
	if t.fileType == file.TypeJSON {
		return data, nil
	}

	if t.fileType != file.TypeEnv {
		return []byte{}, fmt.Errorf("transformer does not support %s files", t.fileType)
	}

	params, err := dotenv.Parse(bytes.NewReader(data))
	if err != nil {
		return data, err
	}

	m := map[string]interface{}{}
	for k, v := range params {

		var value interface{}
		if b, err := strconv.ParseBool(v); err == nil {
			value = b
		} else if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			value = json.Number(v)
		} else if _, err := strconv.ParseFloat(v, 64); err == nil {
			value = v
		} else {
			value = v
		}

		m[k] = value
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return b, err
	}

	return b, nil
}
