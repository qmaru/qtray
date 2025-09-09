package utils

import (
	"embed"
	"encoding/json"
	"os"

	"github.com/getlantern/systray"
)

type Config struct {
	Title   string  `json:"title"`
	Process Process `json:"process"`
	Admin   bool    `json:"admin"`
}

type Process struct {
	Name string   `json:"name"`
	Path string   `json:"path"`
	Args []string `json:"args"`
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

func CreateTrayMenu(title string) (mainItem *systray.MenuItem, quitItem *systray.MenuItem) {
	mainItem = systray.AddMenuItem(title, "")
	mainItem.Disable()
	quitItem = systray.AddMenuItem("exit", "exit")
	return
}
