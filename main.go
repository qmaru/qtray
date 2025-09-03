//go:generate goversioninfo -64
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"qtray/utils"

	"github.com/getlantern/systray"
)

//go:embed app.ico
var iconFS embed.FS

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

func main() {
	iconData, err := iconFS.ReadFile("icon.ico")
	if err != nil {
		log.Fatal("load icon error: ", err)
	}

	config, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("load config error: ", err)
	}

	var trayConfig Config
	if err := json.Unmarshal(config, &trayConfig); err != nil {
		log.Fatal("unmarshal config error: ", err)
	}

	if trayConfig.Process.Name == "" || trayConfig.Process.Path == "" {
		log.Fatal("config error: process name or path is empty")
	}

	systray.Run(onReady(trayConfig, iconData), onExit)
}

func onReady(config Config, iconData []byte) func() {
	return func() {
		utils.InitTray(config.Title, iconData, func() {
			mInfo := systray.AddMenuItem(config.Title, "")
			mInfo.Disable()
			mQuit := systray.AddMenuItem("exit", "exit")

			if utils.IsProcessRunning(config.Process.Name) {
				showErrorAndQuit(fmt.Sprintf("%s is running", config.Process.Name))
				return
			}

			if config.Admin {
				if !utils.IsAdmin() {
					utils.ShowMsgBox("run as administrator")
					utils.AutoElevateSelf()
					systray.Quit()
					return
				}
			}

			cmd, waitCh, err := RunProcess(config.Process.Name, config.Process.Path, config.Process.Args)
			if err != nil {
				showErrorAndQuit(fmt.Sprintf("start %s failed: %v", config.Process.Name, err))
				return
			}

			go func() {
				<-mQuit.ClickedCh
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				systray.Quit()
			}()

			go func() {
				waitErr := <-waitCh
				if waitErr != nil {
					utils.ShowMsgBox(fmt.Sprintf("process exited with error: %v", waitErr))
				}
			}()

		}, onExit)
	}
}

func onExit() {
	log.Println("exit")
}

func RunProcess(name, path string, args []string) (*exec.Cmd, <-chan error, error) {
	exePath := filepath.Join(path, name)
	exePath = os.ExpandEnv(exePath)
	if !filepath.IsAbs(exePath) {
		exePath, _ = filepath.Abs(exePath)
	}

	if _, err := os.Stat(exePath); err != nil {
		return nil, nil, fmt.Errorf("executable %s not found in %s", name, path)
	}

	replacedArgs := make([]string, len(args))
	for i, v := range args {
		replacedArgs[i] = os.ExpandEnv(v)
	}

	cmd := exec.Command(exePath, replacedArgs...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	err := cmd.Start()
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

func showErrorAndQuit(msg string) {
	utils.ShowMsgBox(msg)
	systray.Quit()
}
