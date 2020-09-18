package events

import (
	"bytes"
	"config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// DiscordWebhookRequest defines the format of a webhook request
type DiscordWebhookRequest struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
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
		Value:  config.InstanceName,
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

	request, err := http.NewRequest("POST", config.WebhooksEndpoint, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf(string(body))
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
