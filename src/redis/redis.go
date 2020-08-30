package redis

import (
	"config"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

// Connect actually connects the redis client if enabled
func Connect() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})
}
