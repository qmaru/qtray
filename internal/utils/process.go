package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-ps"
)

func IsProcessRunning(name string) bool {
	processes, err := ps.Processes()
	if err != nil {
		return false
	}

	for _, p := range processes {
		executable := p.Executable()

		if executable == name {
			return true
		}

		if filepath.Base(executable) == name {
			return true
		}

		if filepath.Base(executable) == filepath.Base(name) {
			return true
		}
	}

	return false
}

func BuildExePath(name, path string) (string, error) {
	var exePath string

	if path == "" {
		exePath = name
	} else if filepath.IsAbs(path) {
		exePath = path
	} else {
		exePath = filepath.Join(path, name)
		exePath = os.ExpandEnv(exePath)
		if !filepath.IsAbs(exePath) {
			exePath, _ = filepath.Abs(exePath)
		}
	}

	if path != "" && !filepath.IsAbs(path) {
		if _, err := os.Stat(exePath); err != nil {
			return "", fmt.Errorf("executable %s not found in %s", name, path)
		}
	}

	return exePath, nil
}

func ExpandArgs(args []string) []string {
	replacedArgs := make([]string, len(args))
	for i, v := range args {
		replacedArgs[i] = os.ExpandEnv(v)
	}
	return replacedArgs
}
