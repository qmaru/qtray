//go:build darwin

package utils

import "os/exec"

func openTarget(target string) error {
	return exec.Command("open", target).Start()
}
