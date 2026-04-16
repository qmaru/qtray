package utils

import (
	"embed"
	"encoding/json"
	"os"

	"github.com/getlantern/systray"
)

type Process struct {
	Name string   `json:"name"`
	Path string   `json:"path"`
	Args []string `json:"args"`
}

type Open struct {
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
	Target  string `json:"target"`
}

type Config struct {
	Title   string  `json:"title"`
	Process Process `json:"process"`
	Admin   bool    `json:"admin"`
	Open    Open    `json:"open"`
}

func LoadConfig(c string) (*Config, error) {
	config, err := os.ReadFile(c)
	if err != nil {
		return nil, err
	}

	var trayConfig Config
	if err := json.Unmarshal(config, &trayConfig); err != nil {
		return nil, err
	}
	return &trayConfig, nil
}

func LoadIcon(fs embed.FS, iconFile string) ([]byte, error) {
	iconData, err := fs.ReadFile(iconFile)
	if err != nil {
		return nil, err
	}
	return iconData, nil
}

func NewTrayMenu(title string) *TrayMenu {
	titleItem := systray.AddMenuItem(title, "")
	titleItem.Disable()

	return &TrayMenu{
		title:       title,
		titleItem:   titleItem,
		customItems: make([]*systray.MenuItem, 0),
	}
}

func (tm *TrayMenu) Finalize() {
	if tm.quitItem != nil {
		return
	}
	tm.quitItem = systray.AddMenuItem("exit", "exit")
}

func (tm *TrayMenu) AddItem(title string, tooltip string) *systray.MenuItem {
	item := systray.AddMenuItem(title, tooltip)
	tm.customItems = append(tm.customItems, item)
	return item
}

func (tm *TrayMenu) AddOpenItem(title string, tooltip string, opener Opener) *systray.MenuItem {
	item := systray.AddMenuItem(title, tooltip)
	tm.customItems = append(tm.customItems, item)
	go func() {
		for range item.ClickedCh {
			if target := opener(); target != "" {
				openTarget(target)
			}
		}
	}()
	return item
}

func (tm *TrayMenu) GetTitleItem() *systray.MenuItem {
	return tm.titleItem
}

func (tm *TrayMenu) GetQuitItem() *systray.MenuItem {
	return tm.quitItem
}

func (tm *TrayMenu) GetCustomItems() []*systray.MenuItem {
	return tm.customItems
}
