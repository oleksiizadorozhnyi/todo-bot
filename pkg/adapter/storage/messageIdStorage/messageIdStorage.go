package messageIdStorage

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

func (m *MessageIdStorage) SaveDataToRedis() {
	for _, ch := range ch {
		close(ch)
	}
	encodedData, err := json.Marshal(ch)
	if err != nil {
		zap.L().Error("SaveDataToRedis()", zap.Error(err))
		return
	}
	err = m.client.Set(context.Background(), "mydata", encodedData, 0).Err()
	if err != nil {
		zap.L().Error("SaveDataToRedis()", zap.Error(err))
		return
	}
	return
}

func (m *MessageIdStorage) RestoreDataFromRedis() {
	val, err := m.client.Get(context.Background(), "mydata").Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		zap.L().Error(" RestoreDataFromRedis()", zap.Error(err))
		return
	}
	err = json.Unmarshal([]byte(val), &ch)
	if err != nil {
		zap.L().Error(" RestoreDataFromRedis()", zap.Error(err))
		return
	}
	return
}

func Set(chatID int64, messageID int) {
	_, exists := ch[chatID]
	if exists {
		ch[chatID] <- messageID
	} else {
		ch[chatID] = make(chan int, 100)
		ch[chatID] <- messageID
	}

}

func Get(chatID int64) []int {
	_, exists := ch[chatID]
	var messageIDs []int
	if exists {
		close(ch[chatID])
		for i := range ch[chatID] {
			messageIDs = append(messageIDs, i)
		}
		delete(ch, chatID)
	}
	return messageIDs
}
