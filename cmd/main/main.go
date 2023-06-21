package main

import (
	"github.com/caarlos0/env/v8"
	"go.uber.org/zap"
	"log"
	"telegramBot/internal/config"
	"telegramBot/pkg/adapter/storage"
	"telegramBot/pkg/adapter/storage/messageIdStorage"
	"telegramBot/pkg/adapter/todo-bot"
	"telegramBot/pkg/telegramApi"
)

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	logger := l.Sugar()
	defer l.Sync()

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal(err)
	}
	database := storage.New()
	logic := todo_bot.New(database)
	redis := messageIdStorage.New()
	bot := telegramApi.New(logic, cfg.Token, redis)
	err = bot.Run()
	if err != nil {
		logger.Fatal(err)
	}
}
