package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func InitRedis() (*redis.Client, error) {

	redisAddr := os.Getenv("")
	redisPassword := os.Getenv("")
	redisDB, err := strconv.Atoi(os.Getenv(""))
	if err != nil {
		fmt.Printf("Invalid REDIS_DB value in .env: %v", err)
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	_, err = client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v", err)
		return nil, err
	}

	fmt.Println("Connected to Redis successfully")
	return client, nil
}
