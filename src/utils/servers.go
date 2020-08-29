package utils

import (
	"log"
	"os"
	"path"
	"strings"
)

// CreateMissingServers creates missing servers from MINECRAFT_SERVERS_TO_CREATE
func CreateMissingServers() {
	serversToCreate := strings.Split(MinecraftServersToCreate, ";")

	for _, serverToCreate := range serversToCreate {
		serverDirectoryName := strings.TrimSpace(serverToCreate)
		serverDirectoryPath := path.Join(MinecraftServersDirectory, serverDirectoryName)
		err := os.MkdirAll(serverDirectoryPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}
