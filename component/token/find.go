package token

import (
	"regexp"
	"strings"
)

const token = `\${(.*?)}`

// RemoteKey ...
type RemoteKey struct {
	Key   string
	Field string
}

func (k RemoteKey) String() string {
	return k.Key
}

func newRemoteKey(fileKey string) RemoteKey {
	if strings.Contains(fileKey, "::") {
		parts := strings.Split(fileKey, "::")

		return RemoteKey{
			Key:   parts[0],
			Field: parts[1],
		}
	}

	return RemoteKey{Key: fileKey}
}

// Find ...
func Find(data []byte) map[string]RemoteKey {
	re := regexp.MustCompile(token)

	matches := re.FindAllSubmatch(data, -1)

	tokens := map[string]RemoteKey{}
	for _, m := range matches {
		fileKey := string(m[1])

		tokens[fileKey] = newRemoteKey(fileKey)
	}

	return tokens
}
