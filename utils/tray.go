package utils

import (
	"github.com/getlantern/systray"
)

func InitTray(name string, iconData []byte, onReady func(), onExit func()) {
	systray.SetIcon(iconData)
	systray.SetTooltip(name)
	onReady()
}
