package utils

import (
	"config"
	"log"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

// RedisConnect actually connects the redis client if enabled
func RedisConnect() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Error connecting to Redis, please check config")
	}
}
