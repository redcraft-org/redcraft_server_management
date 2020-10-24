package rcsm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

// MinecraftServer defines the stats about a server
type MinecraftServer struct {
	name         string
	fullPath     string
	running      bool
	crashed      bool
	restartTries int64
	firstRetry   time.Time
	StartCommand string `json:"start_command"`
	StopCommand  string `json:"stop_command"`
}

var (
	minecraftServers       map[string]MinecraftServer = make(map[string]MinecraftServer)
	minecraftServersLock   sync.Mutex
	healthcheckRunning     bool
	healthcheckRunningLock sync.Mutex
	redisSubscribed        bool
)

// DiscoverServers does a scan to know which servers exists
func DiscoverServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	fileNodes, err := ioutil.ReadDir(MinecraftServersDirectory)
	if err != nil {
		TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not scan servers: %s", err))
		os.Exit(1)
	}

	for _, fileNode := range fileNodes {
		if fileNode.IsDir() {
			serverName := fileNode.Name()
			serverPath := path.Join(MinecraftServersDirectory, serverName)

			if S3Enabled {
				if !SessionExists(serverName) {
					UpdateTemplate(serverName)
				} else {
					TriggerLogEvent("info", serverName, "Not updating template, server is running")
				}
			}

			minecraftServer, err := readConfig(serverPath)
			if err != nil {
				TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not read server config: %s", err))
				continue
			}

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}

	if RedisEnabled && !redisSubscribed {
		ListenForRedisCommands()
		redisSubscribed = true
	}

	TriggerLogEvent("info", "setup", fmt.Sprintf("Found %d server(s)", len(minecraftServers)))
}

// ServerExists returns wether a server exists or not
func ServerExists(serverName string) bool {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	_, exists := minecraftServers[serverName]

	return exists
}

// StartServer starts a server with a specified name
func StartServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", serverName, "Starting server")

	server := minecraftServers[serverName]

	startServer(server)
}

// StopServer stops a server with a specified name
func StopServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", serverName, "Stopping server")

	server := minecraftServers[serverName]

	stopServer(server)
}

// RestartServer restarts a server with a specified name
func RestartServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", serverName, "Restarting server")

	server := minecraftServers[serverName]

	stopServer(server)
	startServer(server)
}

// RunCommandServer restarts a server with a specified name
func RunCommandServer(serverName string, command string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", serverName, fmt.Sprintf("Running command `%s`", command))

	server := minecraftServers[serverName]

	runCommand(server, command)
}

// RunCommandAllServers restarts a server with a specified name
func RunCommandAllServers(command string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", "rcsm", fmt.Sprintf("Running command on all servers `%s`", command))
	for _, server := range minecraftServers {
		runCommand(server, command)
	}
}

// StartAllServers starts all servers
func StartAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", "rcsm", "Starting all servers")

	for _, server := range minecraftServers {
		// Let rcsm rest for a bit
		time.Sleep(time.Second * 3)
		startServer(server)
	}
}

// StopAllServers stops all servers
func StopAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", "rcsm", "Stopping all servers")

	for _, server := range minecraftServers {
		// Let rcsm rest for a bit
		time.Sleep(time.Second * 3)
		stopServer(server)
	}
}

// RestartAllServers starts all servers
func RestartAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	TriggerLogEvent("info", "rcsm", "Restarting all servers")

	for _, server := range minecraftServers {
		// Let rcsm rest for a bit
		time.Sleep(time.Second * 3)
		stopServer(server)
		startServer(server)
	}
}

func startServer(server MinecraftServer) bool {
	serverName := server.name
	isRunning := SessionExists(serverName)
	if isRunning {
		TriggerLogEvent("warn", serverName, "Server already started")
		server.running = true
		server.crashed = false
		minecraftServers[serverName] = server
		return true
	}

	attachCommand, err := SessionCreate(serverName, server.fullPath, server.StartCommand)
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not start: %s", err))
		server.running = isRunning
		server.crashed = !isRunning
	} else {
		TriggerLogEvent("info", serverName, fmt.Sprintf("Starting server, run \"%s\" to see the console", attachCommand))
		server.running = true
		server.crashed = false
	}

	minecraftServers[serverName] = server

	return server.running
}

func stopServer(server MinecraftServer) bool {
	serverName := server.name
	isRunning := SessionExists(serverName)
	if !server.running && !isRunning {
		TriggerLogEvent("warn", serverName, "Server already stopped")
		return true
	}

	err := SessionTerminate(server.name, server.StopCommand, false)
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Error while stopping: %s", err))
		server.running = isRunning
		server.crashed = isRunning
	} else {
		TriggerLogEvent("info", serverName, "Stopping server")
		server.running = false
		server.crashed = false
	}

	minecraftServers[serverName] = server

	return !server.running
}

func runCommand(server MinecraftServer, command string) bool {
	serverName := server.name
	if !server.running && !SessionExists(serverName) {
		TriggerLogEvent("warn", serverName, "Tried to run command on a stopped server")
		return true
	}

	err := SessionRunCommand(serverName, command)
	if err != nil {
		TriggerLogEvent("warn", serverName, fmt.Sprintf("Could not run command: %s", err))
	}

	return !server.running
}
