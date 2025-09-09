package helper

type TrayHelper interface {
	AutoElevateSelf()
	IsAdmin() bool
	ShowMsgBox(msg string, btnType uint) int
}
