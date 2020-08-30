package main

import (
	"config"
	"redis"
	"servers"
)

func main() {
	config.ReadConfig()
	servers.CreateMissingServers()
	servers.Discover()

	if config.RedisEnabled {
		redis.Connect()
	}
}
