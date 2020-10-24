package rcsm

import (
	"context"
	"encoding/json"
)

// RedisAvailable is used to know if redis is ready to receive messages
var RedisAvailable bool

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
		Instance: InstanceName,
		Service:  service,
		Message:  message,
	}

	requestPayload, err := json.Marshal(redisRequest)
	if err != nil {
		return err
	}

	response := RedisClient.Publish(context.TODO(), RedisPubSubChannel, string(requestPayload))
	return response.Err()
}
