package servers

import (
	"config"
	"encoding/json"
	"events"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// CreateMissingServers creates missing servers from MINECRAFT_SERVERS_TO_CREATE
func CreateMissingServers() {
	serversToCreate := strings.Split(config.MinecraftServersToCreate, ";")

	for _, serverToCreate := range serversToCreate {
		serverDirectoryName := strings.TrimSpace(serverToCreate)
		serverDirectoryPath := path.Join(config.MinecraftServersDirectory, serverDirectoryName)

		_, err := os.Stat(serverDirectoryPath)
		if os.IsNotExist(err) {
			err = os.MkdirAll(serverDirectoryPath, os.ModePerm)
			if err != nil {
				events.TriggerLogEvent("severe", serverDirectoryName, fmt.Sprintf("Could not create server: %s", err))
			}
			events.TriggerLogEvent("info", serverDirectoryName, "Created directory for server")
		}
	}
}

func readConfig(serverPath string) (MinecraftServer, error) {
	var minecraftServer MinecraftServer
	configFilePath := path.Join(serverPath, "rcsm_config.json")

	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		initConfig(serverPath)
	}

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		return minecraftServer, err
	}
	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(jsonBytes, &minecraftServer)
	if err != nil {
		return minecraftServer, err
	}

	return minecraftServer, nil
}

func initConfig(serverPath string) {
	// Default start args are from here: https://aikar.co/2018/07/02/tuning-the-jvm-g1gc-garbage-collector-flags-for-minecraft/
	statusTemplate := MinecraftServer{
		StartArgs:   "-Xms6G -Xmx6G -XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 -XX:G1HeapRegionSize=8M -XX:G1ReservePercent=20 -XX:G1HeapWastePercent=5 -XX:G1MixedGCCountTarget=4 -XX:InitiatingHeapOccupancyPercent=15 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem -XX:MaxTenuringThreshold=1 -Dusing.aikars.flags=https://mcflags.emc.gs -Daikars.new.flags=true",
		JarName:     "server.jar",
		StopCommand: "stop",
	}

	jsonContents, err := json.MarshalIndent(statusTemplate, "", "    ")
	if err != nil {
		events.TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not serialize default template: %s", err))
		os.Exit(1)
	}

	configFilePath := path.Join(serverPath, "rcsm_config.json")

	err = ioutil.WriteFile(configFilePath, jsonContents, 0644)
	if err != nil {
		events.TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not save default template: %s", err))
		os.Exit(1)
	}

	events.TriggerLogEvent("info", "setup", fmt.Sprintf("Saved default template at %s", configFilePath))
}
