//go:build darwin
// +build darwin

package helper

type MacOSTrayHelper struct{}

var _ TrayHelper = (*MacOSTrayHelper)(nil)

func (t *MacOSTrayHelper) IsAdmin() bool {
	return false
}

func (t *MacOSTrayHelper) ShowMsgBox(msg string, btnType uint) int {
	return 0
}

func (t *MacOSTrayHelper) AutoElevateSelf() {
}

func NewTray() *MacOSTrayHelper {
	return &MacOSTrayHelper{}
}
