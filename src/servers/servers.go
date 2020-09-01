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
)

// Discover does a scan to know which servers exists
func Discover() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	fileNodes, err := ioutil.ReadDir(config.MinecraftServersDirectory)
	if err != nil {
		events.TriggerLogEvent(config.InstanceName, "fatal", "setup", fmt.Sprintf("Could not scan servers: %s", err))
		os.Exit(1)
	}

	for _, fileNode := range fileNodes {
		if fileNode.IsDir() {
			serverName := fileNode.Name()
			serverPath := path.Join(config.MinecraftServersDirectory, serverName)

			if !config.S3Enabled {
				if !tmux.SessionExists(serverName) {
					UpdateTemplate(serverName)
				} else {
					events.TriggerLogEvent(config.InstanceName, "info", serverName, "Not updating template, server is running")
				}
			}

			minecraftServer, err := readConfig(serverPath)
			if err != nil {
				events.TriggerLogEvent(config.InstanceName, "severe", serverName, fmt.Sprintf("Could not read server config: %s", err))
				continue
			}

			minecraftServer.name = serverName
			minecraftServer.fullPath = serverPath

			minecraftServers[serverName] = minecraftServer
		}
	}

	events.TriggerLogEvent(config.InstanceName, "info", "setup", fmt.Sprintf("Found %d server(s)", len(minecraftServers)))
}

// StartServer starts a server with a specified name
func StartServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent(config.InstanceName, "info", serverName, "Starting server")

	server := minecraftServers[serverName]

	startServer(server)
}

// StopServer stops a server with a specified name
func StopServer(serverName string) {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent(config.InstanceName, "info", serverName, "Stopping server")

	server := minecraftServers[serverName]

	stopServer(server)
}

// StartAllServers starts all servers
func StartAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent(config.InstanceName, "info", "rcsm", "Starting all servers")

	for _, server := range minecraftServers {
		startServer(server)
	}
}

// StopAllServers stops all servers
func StopAllServers() {
	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	events.TriggerLogEvent(config.InstanceName, "info", "rcsm", "Stopping all servers")

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
	healthcheckRunningLock.Lock()
	defer healthcheckRunningLock.Unlock()
	if healthcheckRunning {
		return
	}
	healthcheckRunningLock.Unlock()

	// Acquire lock on minecraftServers
	minecraftServersLock.Lock()
	defer minecraftServersLock.Unlock()

	for _, server := range minecraftServers {
		serverName := server.name
		if server.running && !tmux.SessionExists(serverName) {
			crashTimeout, err := time.ParseDuration(fmt.Sprintf("%ds", config.AutoRestartCrashTimeoutSec))
			if err != nil {
				events.TriggerLogEvent(config.InstanceName, "severe", "healthcheck", fmt.Sprintf("Could not parse timeout: %s", err))
			} else if time.Now().Add(-crashTimeout).Before(server.firstRetry) {
				server.restartTries = 0
			}

			// We want to log an automated restart to avoid bootloops
			if server.restartTries == 0 {
				server.firstRetry = time.Now()
			}
			server.restartTries++

			if server.restartTries > config.AutoRestartCrashMaxTries {
				events.TriggerLogEvent(config.InstanceName, "severe", serverName, "Server crash bootloop")
				server.running = false
				server.crashed = true
			} else {
				events.TriggerLogEvent(config.InstanceName, "info", serverName, "Server is stopped, restarting")
				if !config.S3Enabled {
					UpdateTemplate(serverName)
				}
				startServer(server)
			}
			minecraftServers[serverName] = server
		}
	}

	healthcheckRunningLock.Lock()
	healthcheckRunning = false
	// No need to unlock, the defer at the top will do it
}

func startServer(server MinecraftServer) bool {
	serverName := server.name
	if tmux.SessionExists(serverName) {
		events.TriggerLogEvent(config.InstanceName, "info", serverName, "Server already started")
		return true
	}

	attachCommand, err := tmux.SessionCreate(serverName, server.fullPath, server.StartArgs, server.JarName)
	if err != nil {
		events.TriggerLogEvent(config.InstanceName, "severe", serverName, fmt.Sprintf("Could not start: %s", err))
		isRunning := tmux.SessionExists(serverName)
		server.running = isRunning
		server.crashed = !isRunning
	} else {
		events.TriggerLogEvent(config.InstanceName, "info", serverName, fmt.Sprintf("Starting server, run \"%s\" to see the console", attachCommand))
		server.running = true
		server.crashed = false
	}

	minecraftServers[serverName] = server

	return server.running
}

func stopServer(server MinecraftServer) bool {
	serverName := server.name
	if !server.running && !tmux.SessionExists(serverName) {
		events.TriggerLogEvent(config.InstanceName, "info", serverName, "Server already stopped")
		return true
	}

	err := tmux.SessionTerminate(server.name, server.StopCommand, false)
	if err != nil {
		events.TriggerLogEvent(config.InstanceName, "severe", serverName, fmt.Sprintf("Error while stopping: %s", err))
		isRunning := tmux.SessionExists(server.name)
		server.running = isRunning
		server.crashed = isRunning
	} else {
		events.TriggerLogEvent(config.InstanceName, "info", serverName, "Stopping server")
		server.running = false
		server.crashed = false
	}

	minecraftServers[server.name] = server

	return !server.running
}
