package main

import (
	"utils"
)

func main() {
	utils.ReadConfig()
	utils.CreateMissingServers()

	if utils.RedisEnabled {
		utils.RedisConnect()
	}
}
