package main

import (
	"utils"
)

func main() {
	utils.ReadConfig()
	if utils.RedisEnabled {
		utils.RedisConnect()
	}
}
