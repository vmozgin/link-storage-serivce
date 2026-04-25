package cache

import (
	"context"
	"link-storage-service/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(cfg config.Redis) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisCache{Client: client}, nil
}

func (c *RedisCache) Set(ctx context.Context, shortCode string, url string) error {
	if err := c.Client.Set(ctx, "link:"+shortCode, url, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) Get(ctx context.Context, shortCode string) (string, error) {
	link, err := c.Client.Get(ctx, "link:"+shortCode).Result()
	if err != nil {
		return "", err
	}
	return link, nil
}

func (c *RedisCache) Delete(ctx context.Context, shortCode string) error {
	return c.Client.Del(ctx, "link:"+shortCode).Err()
}
