package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"telegramBot/internal/config"
)

type Cache struct {
	client *redis.Client
}

func New(cfg config.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.AddrRedis,
		Password: cfg.PasswordRedis,
		DB:       0,
	})
	return &Cache{
		client: client,
	}
}

func (m *Cache) Set(ctx context.Context, chatID int64, messageID int) error {
	err := m.client.RPush(ctx, strconv.FormatInt(chatID, 10), strconv.Itoa(messageID)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (m *Cache) Get(ctx context.Context, chatID int64) ([]int, error) {
	key := strconv.FormatInt(chatID, 10)
	exists, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists != 0 {
		values, err := m.client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			return nil, err
		}

		result := make([]int, len(values))
		for i, value := range values {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			result[i] = int(intValue)
		}
		m.client.Del(ctx, key)
		return result, nil
	} else {
		return nil, nil
	}
}
