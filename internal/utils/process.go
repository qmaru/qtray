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
		if p.Executable() == name {
			return true
		}
	}

	return false
}

func BuildExePath(name, path string) (string, error) {
	exePath := filepath.Join(path, name)
	exePath = os.ExpandEnv(exePath)
	if !filepath.IsAbs(exePath) {
		exePath, _ = filepath.Abs(exePath)
	}
	if _, err := os.Stat(exePath); err != nil {
		return "", fmt.Errorf("executable %s not found in %s", name, path)
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
