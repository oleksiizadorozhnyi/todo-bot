package messageIdStorage

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
)

var ch = make(map[int64]chan int)

type MessageIdStorage struct {
	client *redis.Client
}

func New() *MessageIdStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return &MessageIdStorage{
		client: client,
	}
}

func (m *MessageIdStorage) Set(chatID int64, messageID int) {
	key := fmt.Sprintf("channel:%d", chatID)
	err := m.client.RPush(context.Background(), key, strconv.Itoa(messageID)).Err()
	if err != nil {
		zap.L().Error("Set()", zap.Error(err))
		return
	}
	return
}

func (m *MessageIdStorage) Get(chatID int64) []int {
	key := fmt.Sprintf("channel:%d", chatID)
	exists, err := m.client.Exists(context.Background(), key).Result()
	if err != nil {
		zap.L().Error("Set()", zap.Error(err))
		return nil
	}
	if exists != 0 {
		values, err := m.client.LRange(context.Background(), key, 0, -1).Result()
		if err != nil {
			zap.L().Error("Get()", zap.Error(err))
			return nil
		}

		result := make([]int, len(values))
		for i, value := range values {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				zap.L().Error("Get()", zap.Error(err))
				return nil
			}
			result[i] = int(intValue)
		}
		return result
	} else {
		return nil
	}
}
