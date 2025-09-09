//go:build windows
// +build windows

//go:generate goversioninfo -64 -o resource_windows.syso build/windows/versioninfo.json
package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"qtray/internal/helper"
	"qtray/internal/utils"

	"github.com/getlantern/systray"
)

//go:embed build/windows/app.ico
var iconFS embed.FS

var currentCmd *exec.Cmd
var tray = helper.NewTray()
var config *utils.Config
var iconData []byte

func init() {
	c, err := utils.LoadConfig("config.json")
	if err != nil {
		tray.ShowMsgBox("config.json not found", helper.WIN_MB_OK)
		return
	}

	i, err := utils.LoadIcon(iconFS, "build/windows/app.ico")
	if err != nil {
		tray.ShowMsgBox("load icon failed", helper.WIN_MB_OK)
		return
	}

	config = c
	iconData = i
}

func main() {
	if config == nil {
		tray.ShowMsgBox("config is not loaded", helper.WIN_MB_OK)
		return
	}

	if config.Process.Name == "" {
		tray.ShowMsgBox("process name is required", helper.WIN_MB_OK)
		return
	}

	if utils.IsProcessRunning(config.Process.Name) {
		tray.ShowMsgBox(fmt.Sprintf("%s is running", config.Process.Name), helper.WIN_MB_OK)
		return
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip(config.Title)
	_, mQuit := utils.CreateTrayMenu(config.Title)

	if config.Admin {
		if !tray.IsAdmin() {
			result := tray.ShowMsgBox("Please run as administrator", helper.WIN_MB_OKCANCEL)
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
		tray.ShowMsgBox(fmt.Sprintf("start %s failed: %v", config.Process.Name, err), helper.WIN_MB_OK)
		systray.Quit()
		return
	}

	currentCmd = cmd

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	go func() {
		waitErr := <-waitCh
		if waitErr != nil {
			tray.ShowMsgBox(fmt.Sprintf("process exited with error: %v", waitErr), helper.WIN_MB_OK)
		}
		systray.Quit()
	}()
}

func onExit() {
	if currentCmd != nil && currentCmd.Process != nil {
		_ = currentCmd.Process.Signal(syscall.SIGTERM)

		done := make(chan error, 1)
		go func() {
			done <- currentCmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = currentCmd.Process.Kill()
		}
	}
}

func RunProcess(name, path string, args []string) (*exec.Cmd, <-chan error, error) {
	exePath, err := utils.BuildExePath(name, path)
	if err != nil {
		return nil, nil, err
	}
	replacedArgs := utils.ExpandArgs(args)

	cmd := exec.Command(exePath, replacedArgs...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

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
