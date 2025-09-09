//go:build darwin
// +build darwin

package helper

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type MacOSTrayHelper struct{}

var _ TrayHelper = (*MacOSTrayHelper)(nil)

// dialog
func dialog(msg, title string, buttons []string, defaultBtn string) (string, error) {
	btns := "\"" + strings.Join(buttons, "\", \"") + "\""
	script := "display dialog \"" + msg + "\" with title \"" + title + "\" buttons {" + btns + "} default button \"" + defaultBtn + "\""
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (t *MacOSTrayHelper) IsAdmin() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	uid := strings.TrimSpace(string(output))
	return uid == "0"
}

func (t *MacOSTrayHelper) ShowMsgBox(msg string, btnType uint) int {
	title := "Message"
	var buttons []string
	var defaultBtn string

	switch btnType {
	case 1:
		buttons = []string{"OK", "Cancel"}
		defaultBtn = "OK"
	case 2:
		buttons = []string{"Yes", "No"}
		defaultBtn = "Yes"
	default:
		buttons = []string{"OK"}
		defaultBtn = "OK"
	}

	result, err := dialog(msg, title, buttons, defaultBtn)
	if err != nil {
		return -1 // error
	}

	switch btnType {
	case 1:
		if strings.Contains(result, "Cancel") {
			return 1 // Cancel
		}
		return 0 // OK
	case 2:
		if strings.Contains(result, "No") {
			return 1 // No
		}
		return 0 // Yes
	}
	return 0 // OK
}

func (t *MacOSTrayHelper) AutoElevateSelf() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, exePath)
	exec.Command("osascript", "-e", script).Start()
}

func NewTray() *MacOSTrayHelper {
	return &MacOSTrayHelper{}
}
