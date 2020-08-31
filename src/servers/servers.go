package servers

import (
	"config"
	"io/ioutil"
	"log"
	"path"
	"sync"
	"time"
	"tmux"
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

			if !tmux.SessionExists(serverName) {
				UpdateTemplate(serverName)
			} else {
				log.Printf("Not updating template for %s, server is running", serverName)
			}

			minecraftServer := readConfig(serverPath)

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}

	log.Printf("Found %d server(s)", len(minecraftServers))
}

// StartServer starts a server with a specified name
func StartServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	server := minecraftServers[serverName]

	startServer(server)
}

// StopServer stops a server with a specified name
func StopServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	log.Printf("Stopping all servers")

	server := minecraftServers[serverName]

	stopServer(server)
}

// StartAllServers starts all servers
func StartAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	log.Printf("Starting all servers")

	for _, server := range minecraftServers {
		startServer(server)
	}
}

// StopAllServers stops all servers
func StopAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	log.Printf("Stopping all servers")

	for _, server := range minecraftServers {
		stopServer(server)
	}
}

// StartHealthCheck starts a task to check that servers are still running
func StartHealthCheck() {
	ticker := time.NewTicker(time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				runHealthCheck()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func runHealthCheck() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	for _, server := range minecraftServers {
		serverName := server.name
		if server.running && !tmux.SessionExists(serverName) {
			log.Printf("Server %s is stopped, restarting", serverName)
			UpdateTemplate(serverName)
			startServer(server)
		}
	}
}

func startServer(server MinecraftServer) bool {
	serverName := server.name
	if server.running || tmux.SessionExists(serverName) {
		log.Printf("%s is already started", serverName)
		return true
	}

	attachCommand, err := tmux.SessionCreate(serverName, server.fullPath, server.StartArgs, server.JarName)
	if err != nil {
		log.Printf("Could not start %s: %s", serverName, err)
		isRunning := tmux.SessionExists(serverName)
		server.running = isRunning
		server.crashed = !isRunning
	} else {
		log.Printf("Starting server %s, run \"%s\" to see the console", serverName, attachCommand)
		server.running = true
		server.crashed = false
	}

	minecraftServers[serverName] = server

	return server.running
}

func stopServer(server MinecraftServer) bool {
	serverName := server.name
	if !server.running && !tmux.SessionExists(serverName) {
		log.Printf("%s is already stopped", serverName)
		return true
	}

	err := tmux.SessionTerminate(server.name, server.StopCommand, false)
	if err != nil {
		log.Printf("Error while stopping %s: %s", server.name, err)
		isRunning := tmux.SessionExists(server.name)
		server.running = isRunning
		server.crashed = isRunning
	} else {
		log.Printf("Stopping server %s", server.name)
		server.running = false
		server.crashed = false
	}

	minecraftServers[server.name] = server

	return !server.running
}
