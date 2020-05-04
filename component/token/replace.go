package token

import (
	"fmt"
	"regexp"
)

const replaceToken = `\${%s?}`

// Replace ...
func Replace(m map[string]string, data []byte) []byte {

	for k, v := range m {
		re := regexp.MustCompile(fmt.Sprintf(replaceToken, regexp.QuoteMeta(k)))

		data = re.ReplaceAll(data, []byte(v))
	}

	return data
}
