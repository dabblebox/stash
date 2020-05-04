package output

import (
	"bytes"
	"encoding/json"
	"fmt"

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

	params, err := dotenv.Parse(bytes.NewReader(data), true)
	if err != nil {
		return data, err
	}

	b, err := json.MarshalIndent(params, "", "    ")
	if err != nil {
		return b, err
	}

	return b, nil
}
