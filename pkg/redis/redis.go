package redis

import (
	"github.com/go-redis/redis/v8"
	"tarkib.uz/config"
)

func NewRedisDB(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: "",
		DB:       0,
	})
	return rdb, nil
}
