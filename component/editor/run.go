package editor

import (
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "vim"

// Run ...
func Run(filePath string) error {

	editor, found := os.LookupEnv("EDITOR")
	if !found {
		editor = defaultEditor
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, resolveEditorArguments(executable, filePath)...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func resolveEditorArguments(executable string, filename string) []string {
	args := []string{filename}

	if strings.Contains(executable, "code") {
		args = append([]string{"--wait"}, args...)
	}

	return args
}
