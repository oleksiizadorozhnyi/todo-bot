package main

import (
	"context"
	"github.com/caarlos0/env/v8"
	"go.uber.org/zap"
	"log"
	"telegramBot/internal/config"
	"telegramBot/pkg/adapter/api/telegram"
	"telegramBot/pkg/adapter/cache/redis"
	"telegramBot/pkg/adapter/storage/sqlite"
	"telegramBot/pkg/adapter/todobot"
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
	storage := sqlite.New()
	logic := todobot.New(storage)
	cache := redis.New(cfg)
	bot := telegram.New(logic, cfg.Token, cache)
	ctx := context.Background()
	err = bot.Run(ctx)
	if err != nil {
		logger.Fatal(err)
	}
}
