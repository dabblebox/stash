package output

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/file"
)

type TaskDefEnvTransformer struct {
	fileType string
}

func (t TaskDefEnvTransformer) Transform(data []byte) ([]byte, error) {
	if t.fileType != file.TypeEnv {
		return []byte{}, fmt.Errorf("transformer does not support %s files", t.fileType)
	}

	type EnvFormat struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	pairs, err := dotenv.Parse(bytes.NewReader(data))
	if err != nil {
		return data, err
	}

	env := []EnvFormat{}
	for key, value := range pairs {
		env = append(env, EnvFormat{
			Value: value,
			Name:  key,
		})
	}

	b, err := json.MarshalIndent(env, "", "    ")
	if err != nil {
		return b, err
	}

	return b, nil
}
