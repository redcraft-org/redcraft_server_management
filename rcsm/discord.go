package rcsm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// DiscordWebhookRequest defines the format of a webhook request
type DiscordWebhookRequest struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

// DiscordErrorMessage defines the format of a webhook request
type DiscordErrorMessage struct {
	Global     bool   `json:"global"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after"`
}

// DiscordEmbed defines the format of a Discord embed message
type DiscordEmbed struct {
	Color  int            `json:"color"`
	Fields []DiscordField `json:"fields"`
}

// DiscordField defines the format of an embed field
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// SendDiscordWebhook sends a webhook request to Discord
func SendDiscordWebhook(level string, service string, message string) error {
	levelField := DiscordField{
		Name:   "Level",
		Value:  level,
		Inline: true,
	}
	instanceField := DiscordField{
		Name:   "Instance",
		Value:  InstanceName,
		Inline: true,
	}
	serviceField := DiscordField{
		Name:   "Server/Service",
		Value:  service,
		Inline: true,
	}
	messageField := DiscordField{
		Name:   "Message",
		Value:  message,
		Inline: false,
	}

	embedMessage := DiscordEmbed{
		Color:  getColorLevel(strings.ToLower(level)),
		Fields: []DiscordField{levelField, instanceField, serviceField, messageField},
	}

	discordRequest := DiscordWebhookRequest{
		Content: "New event",
		Embeds:  []DiscordEmbed{embedMessage},
	}

	jsonRequest, err := json.Marshal(discordRequest)
	if err != nil {
		return err
	}

	for {
		response, err := http.Post(WebhooksEndpoint, "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			return err
		}
		defer response.Body.Close()

		// Detect if we're getting an error
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			body, _ := ioutil.ReadAll(response.Body)

			var errorMessage DiscordErrorMessage
			err = json.Unmarshal(body, &errorMessage)
			if err != nil {
				return err
			}

			// If it's a rate limit, wait and retry
			if errorMessage.RetryAfter > 0 {
				sleepDuration := time.Duration(errorMessage.RetryAfter) * time.Millisecond
				time.Sleep(sleepDuration)
				continue
			}

			// Error was not related to rate limiting
			return fmt.Errorf(string(body))
		}

		// Code was between 200 and 300, all good
		break
	}

	return nil
}

func getColorLevel(level string) int {
	colors := map[string]int{
		"info":   1499250,
		"warn":   14992650,
		"severe": 16144655,
		"fatal":  16141655,
	}

	color, found := colors[level]

	if !found {
		return 0
	}

	return color
}
