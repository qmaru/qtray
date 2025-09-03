package utils

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type MessageBoxButtons uint

const (
	MB_OK          MessageBoxButtons = 0x0 // Shows "OK" button only
	MB_OKCANCEL    MessageBoxButtons = 0x1 // Shows "OK" and "Cancel" buttons
	MB_YESNOCANCEL MessageBoxButtons = 0x3 // Shows "Yes", "No", and "Cancel" buttons
	MB_YESNO       MessageBoxButtons = 0x4 // Shows "Yes" and "No" buttons
	MB_RETRYCANCEL MessageBoxButtons = 0x5 // Shows "Retry" and "Cancel" buttons
)

func IsAdmin() bool {
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

func ShowMsgBox(msg string, btnType MessageBoxButtons) int {
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

func AutoElevateSelf() {
	exe, err := syscall.UTF16PtrFromString(os.Args[0])
	if err != nil {
		return
	}
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	modShell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExec := modShell32.NewProc("ShellExecuteW")
	_, _, _ = procShellExec.Call(0, uintptr(unsafe.Pointer(verbPtr)), uintptr(unsafe.Pointer(exe)), 0, 0, 1)
}
