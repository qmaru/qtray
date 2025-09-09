//go:build windows
// +build windows

package helper

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WIN_MB_OK          uint = 0x0 // Shows "OK" button only
	WIN_MB_OKCANCEL    uint = 0x1 // Shows "OK" and "Cancel" buttons
	WIN_MB_YESNOCANCEL uint = 0x3 // Shows "Yes", "No", and "Cancel" buttons
	WIN_MB_YESNO       uint = 0x4 // Shows "Yes" and "No" buttons
	WIN_MB_RETRYCANCEL uint = 0x5 // Shows "Retry" and "Cancel" buttons
)

type WindowsTrayHelper struct{}

var _ TrayHelper = (*WindowsTrayHelper)(nil)

func (t *WindowsTrayHelper) IsAdmin() bool {
	token := windows.GetCurrentProcessToken()
	defer token.Close()
	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}
	isMember, err := token.IsMember(adminSID)
	if err != nil {
		return false
	}
	return isMember
}

func (t *WindowsTrayHelper) ShowMsgBox(msg string, btnType uint) int {
	title := "Message"
	msgPtr, _ := syscall.UTF16PtrFromString(msg)
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	modUser32 := syscall.NewLazyDLL("user32.dll")
	procMessageBox := modUser32.NewProc("MessageBoxW")
	ret, _, _ := procMessageBox.Call(0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(btnType))
	return int(ret)
}

func (t *WindowsTrayHelper) AutoElevateSelf() {
	exe, err := syscall.UTF16PtrFromString(os.Args[0])
	if err != nil {
		return
	}
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	modShell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExec := modShell32.NewProc("ShellExecuteW")
	_, _, _ = procShellExec.Call(0, uintptr(unsafe.Pointer(verbPtr)), uintptr(unsafe.Pointer(exe)), 0, 0, 1)
}

func NewTray() *WindowsTrayHelper {
	return &WindowsTrayHelper{}
}
