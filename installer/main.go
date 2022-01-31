package main

import (
	"golang.org/x/sys/windows/registry"
	"io"
	"os"
	"os/exec"
)

func main() {
	startupDir := getConfigDir()
	steamDir := getSteamDir()
	currentDir, _ := os.Getwd()

	oldLocation := currentDir + `\gamestate_integration_discordrpc.cfg`
	newLocation := steamDir + `/gamestate_integration_discordrpc.cfg`
	MoveFile(oldLocation, newLocation)

	oldLocation = currentDir + `\csgo-discord-rpc.exe`
	newLocation = startupDir + `\csgo-discord-rpc.exe`
	MoveFile(oldLocation, newLocation)

	cmnd := exec.Command("csgo-discord-rpc.exe")
	cmnd.Start()
}

func getConfigDir() string {
	configDir, _ := os.UserConfigDir()
	return configDir + `\Microsoft\Windows\Start Menu\Programs\Startup`
}

func getSteamDir() string {
	k, _ := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam\`, registry.QUERY_VALUE)
	defer k.Close()

	steamDir, _, _ := k.GetStringValue("SteamPath")

	return steamDir + `/steamapps/common/Counter-Strike Global Offensive/csgo/cfg`
}

func MoveFile(sourcePath, destPath string) {
	inputFile, _ := os.Open(sourcePath)

	outputFile, _ := os.Create(destPath)
	defer outputFile.Close()

	io.Copy(outputFile, inputFile)
	inputFile.Close()
}
