package output

import (
	"bytes"
	"fmt"
	
	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/file"
)

type ExportTransformer struct {
	fileType string
	literal bool
}

func (t ExportTransformer) Transform(data []byte) ([]byte, error) {
	if t.fileType != file.TypeEnv {
		return []byte{}, fmt.Errorf("transformer does not support %s files", t.fileType)
	}

	params, err := dotenv.Parse(bytes.NewReader(data))
	if err != nil {
		return data, err
	}

	quote := getQuote(t.literal)

	var b bytes.Buffer
	for k, v := range params {
		b.WriteString(fmt.Sprintf("export %s=%s%s%s\n", k, quote, v, quote))
	}

	return b.Bytes(), nil
}

func getQuote(literal bool) string {
	if literal {
		return `'`
	}

	return `"`
}
