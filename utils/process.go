package utils

import (
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
