package dotenv

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	// Pattern for detecting valid line format
	linePattern = `\A\s*(?:export\s+)?([\w\.]+)(?:\s*=\s*|:\s+?)('(?:\'|[^'])*'|"(?:\"|[^"])*"|[^#\n]+)?\s*(?:\s*\#.*)?\z`
)

// Env holds key/value pair of valid environment variable
type Env map[string]string

// Parse is a function to parse line by line any io.Reader supplied and returns the valid Env key/value pair of valid variables.
// It expands the value of a variable from the environment variable but does not set the value to the environment itself.
// This function is returning an error if there are any invalid lines.
func Parse(r io.Reader) (Env, error) {
	env := make(Env)
	scanner := bufio.NewScanner(r)

	i := 1
	bom := string([]byte{239, 187, 191})

	for scanner.Scan() {
		line := scanner.Text()

		if i == 1 {
			line = strings.TrimPrefix(line, bom)
		}

		i++

		err := parseLine(line, env)
		if err != nil {
			return env, err
		}
	}

	return env, nil
}

func parseLine(s string, env Env) error {

	rl := regexp.MustCompile(linePattern)
	rm := rl.FindStringSubmatch(s)

	if len(rm) == 0 {
		return checkFormat(s, env)
	}

	key := rm[1]
	val := rm[2]

	// determine if string has quote prefix
	hdq := strings.HasPrefix(val, `"`)

	// trim whitespace
	val = strings.Trim(val, " ")

	// remove quotes '' or ""
	rq := regexp.MustCompile(`\A(['"])(.*)(['"])\z`)
	val = rq.ReplaceAllString(val, "$2")

	if hdq {
		val = strings.Replace(val, `\n`, "\n", -1)
		val = strings.Replace(val, `\r`, "\r", -1)
	}

	env[key] = val

	return nil
}

func parseExport(st string, env Env) error {
	if strings.HasPrefix(st, "export") {
		vs := strings.SplitN(st, " ", 2)

		if len(vs) > 1 {
			if _, ok := env[vs[1]]; !ok {
				return fmt.Errorf("line `%s` has an unset variable", st)
			}
		}
	}

	return nil
}

func checkFormat(s string, env Env) error {
	st := strings.TrimSpace(s)

	if (st == "") || strings.HasPrefix(st, "#") {
		return nil
	}

	if err := parseExport(st, env); err != nil {
		return err
	}

	return fmt.Errorf("line `%s` doesn't match format", s)
}
