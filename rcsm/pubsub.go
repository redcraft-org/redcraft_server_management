package rcsm

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type callbackFunc func(string, string)

// ChannelListener defines a channel listener with its callback
type ChannelListener struct {
	PubSub   *redis.PubSub
	Channel  string
	Callback callbackFunc
}

// StartRedisListener starts a listener and returns a ChannelListener instance
func StartRedisListener(channel string, callback callbackFunc) (*ChannelListener, error) {
	listener := ChannelListener{
		PubSub:   RedisClient.Subscribe(context.TODO(), channel),
		Channel:  channel,
		Callback: callback,
	}

	// Listen for messages
	go listener.listen()

	return &listener, nil
}

func (listener *ChannelListener) listen() error {
	var channel string
	var payload string

	for {
		msg, err := listener.PubSub.ReceiveTimeout(context.TODO(), time.Second)
		if err != nil {
			// Timeout, ignore
			continue
		}

		channel = ""
		payload = ""

		switch m := msg.(type) {
		case *redis.Subscription:
			continue
		case *redis.Message:
			channel = m.Channel
			payload = m.Payload
		}

		// Process the message
		go listener.Callback(channel, payload)
	}
}
