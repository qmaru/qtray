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

const iconFileName = "app.ico"

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
	config, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("load config error: ", err)
	}

	var trayConfig Config
	if err := json.Unmarshal(config, &trayConfig); err != nil {
		log.Fatal("unmarshal config error: ", err)
	}

	iconData, err := iconFS.ReadFile(iconFileName)
	if err != nil {
		log.Fatal("load icon error: ", err)
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
				utils.ShowMsgBox(fmt.Sprintf("%s is running", config.Process.Name), utils.MB_OK)
				systray.Quit()
				return
			}

			if config.Admin {
				if !utils.IsAdmin() {
					result := utils.ShowMsgBox("Please run as administrator", utils.MB_OKCANCEL)
					if result == 2 {
						systray.Quit()
						return
					}
					utils.AutoElevateSelf()
					systray.Quit()
					return
				}
			}

			cmd, waitCh, err := RunProcess(config.Process.Name, config.Process.Path, config.Process.Args)
			if err != nil {
				utils.ShowMsgBox(fmt.Sprintf("start %s failed: %v", config.Process.Name, err), utils.MB_OK)
				return
			}

			go func() {
				<-mQuit.ClickedCh
				if cmd.Process != nil {
					_ = cmd.Process.Signal(syscall.SIGTERM)
					_ = cmd.Process.Kill()
				}
				systray.Quit()
			}()

			go func() {
				waitErr := <-waitCh
				if waitErr != nil {
					utils.ShowMsgBox(fmt.Sprintf("process exited with error: %v", waitErr), utils.MB_OK)
				}
				systray.Quit()
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
