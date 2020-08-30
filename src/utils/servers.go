package utils

import (
	"config"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

// MinecraftServer defines the stats about a server
type MinecraftServer struct {
	name         string
	fullPath     string
	running      bool
	crashed      bool
	startHistory []int64
	restartTries int
	StartArgs    string `json:"start_args"`
	JarName      string `json:"jar_name"`
	StopCommand  string `json:"stop_command"`
}

var minecraftServers map[string]MinecraftServer = make(map[string]MinecraftServer)
var minecraftServersLock sync.Mutex

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
				log.Fatal("Could not create server: ", err)
			}
			log.Printf("Created directory for server %s", serverToCreate)
		}
	}
}

// DiscoverServers does a scan to know which servers exists
func DiscoverServers() {
	fileNodes, err := ioutil.ReadDir(config.MinecraftServersDirectory)
	if err != nil {
		log.Fatal("Could not scan servers: ", err)
	}

	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	for _, fileNode := range fileNodes {
		if fileNode.IsDir() {
			serverName := fileNode.Name()
			serverPath := path.Join(config.MinecraftServersDirectory, serverName)

			minecraftServer := readConfig(serverPath)

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}
}

func readConfig(serverPath string) MinecraftServer {
	configFilePath := path.Join(serverPath, "rcsm_config.json")

	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		initConfig(serverPath)
	}

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal("Could not read server config: ", err)
	}
	defer jsonFile.Close()

	var minecraftServer MinecraftServer
	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(jsonBytes, &minecraftServer)
	if err != nil {
		log.Fatal("Could not decode server config: ", err)
	}

	log.Printf("Read %s", configFilePath)

	return minecraftServer
}

func initConfig(serverPath string) {
	statusTemplate := MinecraftServer{
		StartArgs:   "-Xmx4G",
		JarName:     "server.jar",
		StopCommand: "stop",
	}

	jsonContents, err := json.MarshalIndent(statusTemplate, "", "    ")
	if err != nil {
		log.Fatal("Could not serialize default template: ", err)
	}

	configFilePath := path.Join(serverPath, "rcsm_config.json")

	err = ioutil.WriteFile(configFilePath, jsonContents, 0644)
	if err != nil {
		log.Fatal("Could not save rcsm_config.json: ", err)
	}

	log.Printf("Created default file %s", configFilePath)
}
