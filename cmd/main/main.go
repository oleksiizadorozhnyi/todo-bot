package main

import (
	"github.com/caarlos0/env/v8"
	"log"
	"telegramBot/internal/config"
	"telegramBot/pkg/adapter/storage"
	"telegramBot/pkg/adapter/todo-bot"
	"telegramBot/pkg/telegramApi"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	database := storage.New()
	logic := todo_bot.New(database)
	bot := telegramApi.New(logic, cfg.Token)
	err := bot.Start()
	if err != nil {
		log.Fatal(err)
	}
}
