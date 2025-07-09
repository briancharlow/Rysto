package redis


import (
	"context"
	"os"
	

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // e.g., redis:6379
		Password: "", // Set if you're using authentication
		DB:       0,
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}
