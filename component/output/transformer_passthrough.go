package output

type PasshroughTransformer struct {
}

func (t PasshroughTransformer) Transform(data []byte) ([]byte, error) {
	return data, nil
}
