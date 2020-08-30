package main

import (
	"config"
	"utils"
)

func main() {
	config.ReadConfig()
	utils.CreateMissingServers()
	utils.DiscoverServers()

	if config.RedisEnabled {
		utils.RedisConnect()
	}
}
