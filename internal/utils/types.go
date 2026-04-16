package utils

import (
	"github.com/getlantern/systray"
)

type Opener func() string

type TrayMenu struct {
	title       string
	titleItem   *systray.MenuItem
	customItems []*systray.MenuItem
	quitItem    *systray.MenuItem
}
