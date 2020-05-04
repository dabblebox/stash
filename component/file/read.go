package file

import (
	"bytes"
	"os"
)

// Read ...
func Read(filePath string) ([]byte, error) {
	data, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(data); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), err
}
