package redis

import (
	"config"
	"events"
	"fmt"

	"gopkg.in/redis.v2"
)

var redisClient *redis.Client

// Connect actually connects the redis client if enabled
func Connect() {
	events.TriggerLogEvent(config.InstanceName, "info", "redis", fmt.Sprintf("Connecting to %s", config.RedisHost))

	redisClient = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})

	response := redisClient.Ping()
	if response.String() != "PING: PONG" {
		events.TriggerLogEvent(config.InstanceName, "severe", "redis", fmt.Sprintf("Error while connecting: %s", response))
	} else {
		events.TriggerLogEvent(config.InstanceName, "info", "redis", "Connected")
	}
}
