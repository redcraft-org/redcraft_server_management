package redis

import (
	"config"
	"context"
	"events"
	"fmt"

	"github.com/go-redis/redis"
)

// RedisClient is the client instance
var RedisClient *redis.Client

// Connect actually connects the redis client if enabled
func Connect() {
	events.TriggerLogEvent("debug", "redis", fmt.Sprintf("Connecting to %s", config.RedisHost))

	RedisClient = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       int(config.RedisDatabase),
	})

	response := RedisClient.Ping(context.TODO())
	if response.String() != "ping: PONG" {
		events.TriggerLogEvent("severe", "redis", fmt.Sprintf("Error while connecting: `%s`", response.String()))
	} else {
		events.TriggerLogEvent("debug", "redis", "Connected")
		events.RedisClient = RedisClient
		events.RedisAvailable = true
	}
}
