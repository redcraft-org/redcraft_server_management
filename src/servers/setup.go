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
				events.TriggerLogEvent(config.InstanceName, "severe", serverDirectoryName, fmt.Sprintf("Could not create server: ", err))
			}
			events.TriggerLogEvent(config.InstanceName, "info", serverDirectoryName, "Created directory for server")
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
	statusTemplate := MinecraftServer{
		StartArgs:   "-Xmx4G",
		JarName:     "server.jar",
		StopCommand: "stop",
	}

	jsonContents, err := json.MarshalIndent(statusTemplate, "", "    ")
	if err != nil {
		events.TriggerLogEvent(config.InstanceName, "severe", "setup", fmt.Sprintf("Could not serialize default template: ", err))
	}

	configFilePath := path.Join(serverPath, "rcsm_config.json")

	err = ioutil.WriteFile(configFilePath, jsonContents, 0644)
	if err != nil {
		events.TriggerLogEvent(config.InstanceName, "severe", "setup", fmt.Sprintf("Could not save default template: ", err))
	}

	events.TriggerLogEvent(config.InstanceName, "info", "setup", fmt.Sprintf("Saved default template at %s", configFilePath))
}
