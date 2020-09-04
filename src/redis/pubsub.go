package redis

import (
	"time"

	"github.com/davidhhuan/go-redis.v2"
)

type callbackFunc func(string, string)

// ChannelListener defines a channel listener with its callback
type ChannelListener struct {
	PubSub   *redis.PubSub
	Channel  string
	Callback callbackFunc
}

// StartListener starts a listener and returns a ChannelListener instance
func StartListener(channel string, callback callbackFunc) (*ChannelListener, error) {
	var err error

	listener := ChannelListener{
		PubSub:   RedisClient.PubSub(),
		Channel:  channel,
		Callback: callback,
	}

	err = listener.PubSub.Subscribe(listener.Channel)
	if err != nil {
		return nil, err
	}

	// Listen for messages
	go listener.listen()

	return &listener, nil
}

func (listener *ChannelListener) listen() error {
	var channel string
	var payload string

	for {
		msg, err := listener.PubSub.ReceiveTimeout(time.Second)
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
		case *redis.PMessage:
			channel = m.Channel
			payload = m.Payload
		}

		// Process the message
		go listener.Callback(channel, payload)
	}
}
