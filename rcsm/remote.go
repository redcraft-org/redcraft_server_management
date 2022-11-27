package rcsm

import (
	"encoding/json"
)

// RedisCommand defines the format of a redis command
type RedisCommand struct {
	Target  string `json:"target"`
	Action  string `json:"action"`
	Content string `json:"content"`
}

// ListenForRedisCommands initializes the listener to listen for redis commands
func ListenForRedisCommands() {
	StartRedisListener(RedisPubSubChannel, parseRedisMessage)
}

func parseRedisMessage(channel string, payload string) {
	var redisCommand RedisCommand
	err := json.Unmarshal([]byte(payload), &redisCommand)
	if err != nil {
		return
	}

	serverName := redisCommand.Target

	if serverName == "*" {
		switch redisCommand.Action {
		case "start":
			StartAllServers()
		case "stop":
			StopAllServers()
		case "restart":
			RestartAllServers()
		case "backup":
			BackupAllServers()
		case "run":
			RunCommandAllServers(redisCommand.Content)
		}
	} else if ServerExists(serverName) {
		switch redisCommand.Action {
		case "start":
			StartServer(serverName)
		case "stop":
			StopServer(serverName)
		case "restart":
			RestartServer(serverName)
		case "backup":
			BackupServer(serverName)
		case "run":
			RunCommandServer(serverName, redisCommand.Content)
		}
	}

}
