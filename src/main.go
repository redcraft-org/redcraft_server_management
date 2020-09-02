package main

import (
	"config"
	"events"
	"fmt"
	"os"
	"os/signal"
	"redis"
	"servers"
	"syscall"
)

func main() {
	initialize()
	waitForQuitSignal()
	stop()
}

func initialize() {
	events.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Starting rcsm (RedCraft Server Manager) v%s", config.Version))

	config.ReadConfig()
	servers.CreateMissingServers()
	servers.Discover()

	if config.RedisEnabled {
		redis.Connect()
	}

	if config.AutoStartOnBoot {
		servers.StartAllServers()
	}

	servers.StartHealthCheck()
}

func stop() {
	events.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Stopping rcsm (RedCraft Server Manager) v%s", config.Version))

	if config.AutoStopOnClose {
		servers.StopAllServers()
	}
}

func waitForQuitSignal() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
