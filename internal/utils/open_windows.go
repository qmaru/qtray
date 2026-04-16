//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

func openTarget(target string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return cmd.Start()
}
