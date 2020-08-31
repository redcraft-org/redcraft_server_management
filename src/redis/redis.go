package redis

import (
	"config"
	"log"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

// Connect actually connects the redis client if enabled
func Connect() {
	log.Printf("Connecting to Redis")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})

	log.Printf("Redis connected")
}
