package main

import (
	"os"
	"log"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/go-redis/redis"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redis_db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redis_db = 0
	}

	redis_client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redis_db,
	})

	pong, err := redis_client.Ping().Result()
	log.Println(pong, err)
}
