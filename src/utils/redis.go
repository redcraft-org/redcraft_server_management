package utils

import (
	"log"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

// RedisConnect actually connects the redis client if enabled
func RedisConnect() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     RedisHost,
		Password: RedisPassword,
		DB:       RedisDatabase,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Error connecting to Redis, please check config")
	}
}
