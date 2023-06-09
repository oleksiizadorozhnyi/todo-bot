package telegramApi

import (
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"os"
	"telegramBot/pkg/adapter/todo-bot"
)

type TelegramApi struct {
	bot     *telego.Bot
	todoBot *todo_bot.TodoBot
}

func New(todoBot *todo_bot.TodoBot, token string) *TelegramApi {
	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &TelegramApi{
		bot:     bot,
		todoBot: todoBot,
	}
}

func (t *TelegramApi) Start() error {
	updates, err := t.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return err
	}
	bh, err := th.NewBotHandler(t.bot, updates)
	if err != nil {
		return err
	}
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!", update.Message.From.FirstName,
		))
	}, th.CommandEqual("start"))
	if err != nil {
		return err
	}

	bh.Handle(func(bot *telego.Bot, update telego.Update) {

		_, err = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			t.todoBot.NewTask(),
		))
	}, th.CommandEqual("createtask"))
	if err != nil {
		return err
	}

	bh.Handle(func(bot *telego.Bot, update telego.Update) {

		_, err = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			t.todoBot.ListOfTasks(),
		))
	}, th.CommandEqual("listOftasks"))
	if err != nil {
		return err
	}

	bh.Handle(func(bot *telego.Bot, update telego.Update) {

		_, err = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			t.todoBot.DeleteTask(),
		))
	}, th.CommandEqual("deletetask"))
	if err != nil {
		return err
	}

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err = bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"Unknown command, use /start or /listOftasks or /createtask  or /deletetask",
		))
	}, th.AnyCommand())
	if err != nil {
		return err
	}

	bh.Start()
	return nil
}
