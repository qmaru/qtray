//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"qtray/internal/helper"
	"qtray/internal/utils"

	"github.com/getlantern/systray"
)

var currentCmd *exec.Cmd
var tray = helper.NewTray()
var config *utils.Config

func init() {
	c, err := utils.LoadConfig("config.json")
	if err != nil {
		tray.ShowMsgBox("config.json not found", 0)
		return
	}

	config = c
}

func main() {
	if config.Process.Name == "" || config.Process.Path == "" {
		tray.ShowMsgBox("process config error", 0)
		return
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	_, mQuit := utils.CreateTrayMenu(config.Title)

	if utils.IsProcessRunning(config.Process.Name) {
		tray.ShowMsgBox(fmt.Sprintf("%s is running", config.Process.Name), 0)
		systray.Quit()
		return
	}

	if config.Admin {
		if !tray.IsAdmin() {
			result := tray.ShowMsgBox("Please run as root", 0)
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
