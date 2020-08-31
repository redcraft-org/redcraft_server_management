package main

import (
	"config"
	"log"
	"redis"
	"servers"
)

func main() {
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

	servers.StopAllServers() // TODO auto shutdown tests
}
