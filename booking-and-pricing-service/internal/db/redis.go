package db

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedisDB(ctx context.Context, addr, password string) (*RedisDB, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Println("connected to redis")
	return &RedisDB{Client: rdb}, nil
}

func (db *RedisDB) Close() error {
	if db.Client != nil {
		return db.Client.Close()
	}
	return nil
}
