package events

import (
	"config"
	"context"
	"encoding/json"

	"github.com/go-redis/redis"
)

// RedisClient will be set by the redis package
var (
	RedisClient    *redis.Client
	RedisAvailable bool
)

// RedisMessage defines the structure of the messages we send on Redis
type RedisMessage struct {
	Level    string `json:"level"`
	Instance string `json:"instance"`
	Service  string `json:"service"`
	Message  string `json:"message"`
}

// SendRedisEvent sends an event on Redis
func SendRedisEvent(level string, service string, message string) error {
	redisRequest := RedisMessage{
		Level:    level,
		Instance: config.InstanceName,
		Service:  service,
		Message:  message,
	}

	requestPayload, err := json.Marshal(redisRequest)
	if err != nil {
		return err
	}

	response := RedisClient.Publish(context.TODO(), config.RedisPubSubChannel, string(requestPayload))
	return response.Err()
}
