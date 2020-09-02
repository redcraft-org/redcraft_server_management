package events

import (
	"log"
	"strings"
)

var instanceName string = "server"

// SetInstanceName is used to set the reported instance name
func SetInstanceName(instanceNameConfig string) {
	instanceName = instanceNameConfig
}

// TriggerLogEvent is the method used to log messages so they can be broadcasted on Redis and Webhooks
func TriggerLogEvent(level string, service string, message string) {
	log.Printf("[%s][%s][%s] %s", instanceName, strings.ToUpper(level), service, message)

	if webhooksEnabled && level != "debug" {
		err := SendDiscordWebhook(level, service, message)
		if err != nil {
			log.Printf("Error while sending webhook: %s", err)
		}
	}
}
