package rcsm

import (
	"log"
	"strings"
)

// TriggerLogEvent is the method used to log messages so they can be broadcasted on Redis and Webhooks
func TriggerLogEvent(level string, service string, message string) {
	level = strings.ToUpper(level)

	log.Printf("[%s][%s][%s] %s", InstanceName, level, service, message)

	if WebhooksEnabled && strings.ToLower(level) != "debug" {
		err := SendDiscordWebhook(level, service, message)
		if err != nil {
			log.Printf("Error while sending webhook: %s", err)
		}
	}

	if RedisAvailable {
		err := SendRedisEvent(level, service, message)
		if err != nil {
			log.Printf("Error while sending Redis pub: %s", err)
		}
	}
}
