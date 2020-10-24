package rcsm

import (
	"fmt"
	"time"
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
		if server.running && !SessionExists(serverName) {
			crashTimeout, err := time.ParseDuration(fmt.Sprintf("%ds", AutoRestartCrashTimeoutSec))
			if err != nil {
				TriggerLogEvent("severe", "healthcheck", fmt.Sprintf("Could not parse timeout: %s", err))
			} else if time.Now().Add(-crashTimeout).After(server.firstRetry) {
				server.restartTries = 0
			}

			// We want to log an automated restart to avoid bootloops
			if server.restartTries == 0 {
				server.firstRetry = time.Now()
			}
			server.restartTries++

			if server.restartTries > AutoRestartCrashMaxTries {
				TriggerLogEvent("severe", serverName, "Server crash bootloop detected")
				server.running = false
				server.crashed = true
			} else {
				TriggerLogEvent("warn", serverName, "Server is stopped, restarting")
				if S3Enabled {
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
