package servers

import (
	"config"
	"events"
	"fmt"
	"time"
	"tmux"
)

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
				events.TriggerLogEvent("severe", "healthcheck", fmt.Sprintf("Could not parse timeout: %s", err))
			} else if time.Now().Add(-crashTimeout).Before(server.firstRetry) {
				server.restartTries = 0
			}

			// We want to log an automated restart to avoid bootloops
			if server.restartTries == 0 {
				server.firstRetry = time.Now()
			}
			server.restartTries++

			if server.restartTries > config.AutoRestartCrashMaxTries {
				events.TriggerLogEvent("severe", serverName, "Server crash bootloop")
				server.running = false
				server.crashed = true
			} else {
				events.TriggerLogEvent("warn", serverName, "Server is stopped, restarting")
				if config.S3Enabled {
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
