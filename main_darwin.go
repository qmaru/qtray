//go:build darwin
// +build darwin

package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"qtray/internal/helper"
	"qtray/internal/utils"

	"github.com/getlantern/systray"
)

//go:embed build/darwin/qtray.app/Contents/Resources/qtray.icns
var iconFS embed.FS

var currentCmd *exec.Cmd
var tray = helper.NewTray()
var config *utils.Config
var iconData []byte

func init() {
	configPath := filepath.Join(filepath.Dir(os.Args[0]), "config.json")
	c, err := utils.LoadConfig(configPath)
	if err != nil {
		tray.ShowMsgBox("config.json not found", 0)
		return
	}

	i, err := utils.LoadIcon(iconFS, "build/darwin/qtray.app/Contents/Resources/qtray.icns")
	if err != nil {
		tray.ShowMsgBox("load icon failed", 0)
		return
	}

	config = c
	iconData = i
}

func main() {
	if config.Process.Name == "" {
		tray.ShowMsgBox("process config error", 0)
		return
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip(config.Title)

	_, mQuit := utils.CreateTrayMenu(config.Title)

	if utils.IsProcessRunning(config.Process.Name) {
		tray.ShowMsgBox(fmt.Sprintf("%s is running", config.Process.Name), 0)
		systray.Quit()
		return
	}

	if config.Admin {
		if !tray.IsAdmin() {
			result := tray.ShowMsgBox("Please run as root", 2)
			if result == 2 {
				systray.Quit()
				return
			}
			tray.AutoElevateSelf()
			systray.Quit()
			return
		}
	}

	cmd, waitCh, err := RunProcess(config.Process.Name, config.Process.Path, config.Process.Args)
	if err != nil {
		tray.ShowMsgBox(fmt.Sprintf("start %s failed: %v", config.Process.Name, err), 0)
	}

	currentCmd = cmd

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	go func() {
		waitErr := <-waitCh
		if waitErr != nil {
			tray.ShowMsgBox(fmt.Sprintf("process exited with error: %v", waitErr), 0)
		}
		systray.Quit()
	}()
}

func onExit() {
	if currentCmd != nil && currentCmd.Process != nil {
		_ = currentCmd.Process.Signal(syscall.SIGTERM)
		_ = currentCmd.Process.Kill()
	}
}

func RunProcess(name, path string, args []string) (*exec.Cmd, <-chan error, error) {
	exePath, err := utils.BuildExePath(name, path)
	if err != nil {
		return nil, nil, err
	}
	replacedArgs := utils.ExpandArgs(args)

	fmt.Printf("Trying to execute: %s\n", exePath)

	cmd := exec.Command(exePath, replacedArgs...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("start %s failed: %v", name, err)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
		close(waitCh)
	}()

	return cmd, waitCh, nil
}
