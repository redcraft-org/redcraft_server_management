package rcsm

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisClient is the client instance
var RedisClient *redis.Client

// RedisConnect actually RedisConnects the redis client if enabled
func RedisConnect() {
	TriggerLogEvent("debug", "redis", fmt.Sprintf("RedisConnecting to %s", RedisHost))

	RedisClient = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     RedisHost,
		Password: RedisPassword,
		DB:       int(RedisDatabase),
	})

	response := RedisClient.Ping(context.TODO())
	if response.String() != "ping: PONG" {
		TriggerLogEvent("severe", "redis", fmt.Sprintf("Error while RedisConnecting: `%s`", response.String()))
	} else {
		TriggerLogEvent("debug", "redis", "RedisConnected")
		RedisAvailable = true
	}
}
