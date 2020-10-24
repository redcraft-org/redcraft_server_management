package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/redcraft-org/redcraft_server_management/rcsm"
)

func main() {
	initialize()
	waitForQuitSignal()
	stop()
}

func initialize() {
	if rcsm.ReadEnvBool("DUMP_VERSION_AND_EXIT", false) {
		fmt.Print(rcsm.Version)
		os.Exit(0)
	}

	rcsm.ReadConfig()

	rcsm.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Starting rcsm (RedCraft Server Manager) v%s", rcsm.Version))

	if rcsm.RedisEnabled {
		rcsm.RedisConnect()
	}

	rcsm.CreateMissingServers()
	rcsm.DiscoverServers()

	if rcsm.AutoStartOnBoot {
		rcsm.StartAllServers()
	}

	if rcsm.AutoRestartCrashEnabled {
		rcsm.StartHealthCheck()
	}

	if rcsm.AutoUpdateEnabled {
		rcsm.StartUpdateChecks()
	}
}

func stop() {
	rcsm.TriggerLogEvent("info", "rcsm", fmt.Sprintf("Stopping rcsm (RedCraft Server Manager) v%s", rcsm.Version))

	if rcsm.AutoStopOnClose {
		rcsm.StopAllServers()
	}
}

func waitForQuitSignal() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
