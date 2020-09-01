package events

import (
	"log"
	"strings"
)

// TriggerLogEvent is the method used to log messages so they can be broadcasted on Redis and Webhooks
func TriggerLogEvent(instanceName string, level string, service string, message string) {
	log.Printf("[%s][%s][%s] %s", instanceName, strings.ToUpper(level), service, message)
}
