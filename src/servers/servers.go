package servers

import (
	"config"
	"io/ioutil"
	"log"
	"path"
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

var (
	minecraftServers     map[string]MinecraftServer = make(map[string]MinecraftServer)
	minecraftServersLock sync.Mutex
)

// Discover does a scan to know which servers exists
func Discover() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	fileNodes, err := ioutil.ReadDir(config.MinecraftServersDirectory)
	if err != nil {
		log.Fatal("Could not scan servers: ", err)
	}

	for _, fileNode := range fileNodes {
		if fileNode.IsDir() {
			serverName := fileNode.Name()
			serverPath := path.Join(config.MinecraftServersDirectory, serverName)

			UpdateTemplate(serverName)

			minecraftServer := readConfig(serverPath)

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}
}
