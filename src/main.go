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
	"update"
)

func main() {
	initialize()
	waitForQuitSignal()
	stop()
}

func initialize() {
	if config.ReadEnvBool("DUMP_VERSION_AND_EXIT", false) {
		fmt.Print(config.Version)
		os.Exit(0)
	}

	config.ReadConfig()

	events.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Starting rcsm (RedCraft Server Manager) v%s", config.Version))

	if config.RedisEnabled {
		redis.Connect()
	}

	servers.CreateMissingServers()
	servers.Discover()

	if config.AutoStartOnBoot {
		servers.StartAllServers()
	}

	if config.AutoRestartCrashEnabled {
		servers.StartHealthCheck()
	}

	if config.AutoUpdateEnabled {
		update.StartUpdateChecks()
	}
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
