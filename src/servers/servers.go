package servers

import (
	"config"
	"events"
	"fmt"
	"io/ioutil"
	"os"
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
	restartTries int64
	firstRetry   time.Time
	StartArgs    string `json:"start_args"`
	JarName      string `json:"jar_name"`
	StopCommand  string `json:"stop_command"`
}

var (
	minecraftServers       map[string]MinecraftServer = make(map[string]MinecraftServer)
	minecraftServersLock   sync.Mutex
	healthcheckRunning     bool
	healthcheckRunningLock sync.Mutex
	redisSubscribed        bool
)

// Discover does a scan to know which servers exists
func Discover() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	fileNodes, err := ioutil.ReadDir(config.MinecraftServersDirectory)
	if err != nil {
		events.TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not scan servers: %s", err))
		os.Exit(1)
	}

	for _, fileNode := range fileNodes {
		if fileNode.IsDir() {
			serverName := fileNode.Name()
			serverPath := path.Join(config.MinecraftServersDirectory, serverName)

			if config.S3Enabled {
				if !tmux.SessionExists(serverName) {
					UpdateTemplate(serverName)
				} else {
					events.TriggerLogEvent("info", serverName, "Not updating template, server is running")
				}
			}

			minecraftServer, err := readConfig(serverPath)
			if err != nil {
				events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not read server config: %s", err))
				continue
			}

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}

	if config.RedisEnabled && !redisSubscribed {
		ListenForRedisCommands()
		redisSubscribed = true
	}

	events.TriggerLogEvent("info", "setup", fmt.Sprintf("Found %d server(s)", len(minecraftServers)))
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

	events.TriggerLogEvent("info", serverName, "Starting server")

	server := minecraftServers[serverName]

	startServer(server)
}

// StopServer stops a server with a specified name
func StopServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent("info", serverName, "Stopping server")

	server := minecraftServers[serverName]

	stopServer(server)
}

// RestartServer restarts a server with a specified name
func RestartServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent("info", serverName, "Restarting server")

	server := minecraftServers[serverName]

	stopServer(server)
	startServer(server)
}

// RunCommandServer restarts a server with a specified name
func RunCommandServer(serverName string, command string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent("info", serverName, fmt.Sprintf("Running command `%s`", command))

	server := minecraftServers[serverName]

	runCommand(server, command)
}

// RunCommandAllServers restarts a server with a specified name
func RunCommandAllServers(command string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Running command on all servers `%s`", command))
	for _, server := range minecraftServers {
		runCommand(server, command)
	}
}

// StartAllServers starts all servers
func StartAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent("info", "rcsm", "Starting all servers")

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

	events.TriggerLogEvent("info", "rcsm", "Stopping all servers")

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

	events.TriggerLogEvent("info", "rcsm", "Restarting all servers")

	for _, server := range minecraftServers {
		// Let rcsm rest for a bit
		time.Sleep(time.Second * 3)
		stopServer(server)
		startServer(server)
	}
}

func startServer(server MinecraftServer) bool {
	serverName := server.name
	isRunning := tmux.SessionExists(serverName)
	if isRunning {
		events.TriggerLogEvent("warn", serverName, "Server already started")
		server.running = true
		server.crashed = false
		minecraftServers[serverName] = server
		return true
	}

	attachCommand, err := tmux.SessionCreate(serverName, server.fullPath, server.StartArgs, server.JarName)
	if err != nil {
		events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not start: %s", err))
		server.running = isRunning
		server.crashed = !isRunning
	} else {
		events.TriggerLogEvent("info", serverName, fmt.Sprintf("Starting server, run \"%s\" to see the console", attachCommand))
		server.running = true
		server.crashed = false
	}

	minecraftServers[serverName] = server

	return server.running
}

func stopServer(server MinecraftServer) bool {
	serverName := server.name
	isRunning := tmux.SessionExists(serverName)
	if !server.running && !isRunning {
		events.TriggerLogEvent("warn", serverName, "Server already stopped")
		return true
	}

	err := tmux.SessionTerminate(server.name, server.StopCommand, false)
	if err != nil {
		events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Error while stopping: %s", err))
		server.running = isRunning
		server.crashed = isRunning
	} else {
		events.TriggerLogEvent("info", serverName, "Stopping server")
		server.running = false
		server.crashed = false
	}

	minecraftServers[server.name] = server

	return !server.running
}

func runCommand(server MinecraftServer, command string) bool {
	serverName := server.name
	if !server.running && !tmux.SessionExists(serverName) {
		events.TriggerLogEvent("warn", serverName, "Tried to run command on a stopped server")
		return true
	}

	err := tmux.SessionRunCommand(serverName, command)
	if err != nil {
		events.TriggerLogEvent("warn", serverName, fmt.Sprintf("Could not run command: %s", err))
	}

	return !server.running
}
