package main

import (
	"config"
	"log"
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
	log.Printf("Starting rcsm (RedCraft Server Manager) v%s", config.Version)

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
	log.Printf("Stopping rcsm (RedCraft Server Manager) v%s", config.Version)

	if config.AutoStopOnClose {
		servers.StopAllServers()
	}
}

func waitForQuitSignal() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
