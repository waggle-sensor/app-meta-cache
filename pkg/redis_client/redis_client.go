package redis_client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

// Redis version comes from the Redis image in Dockerfile
const REDIS_CLIENT_VER = "7.0.4"

type RedisClient struct {
	client *redis.Client
	host   string
}

func NewRedisClient(host string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:6379", host),
		}),
		host: host,
	}
}

func (r *RedisClient) Set(k string, v interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.client.Set(ctx, k, v, 0).Err()
}
