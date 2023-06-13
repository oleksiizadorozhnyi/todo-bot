package telegramApi

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"os"
	todo_bot "telegramBot/pkg/adapter/todo-bot"
	"telegramBot/pkg/model/state"
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
	botUser, err := bot.GetMe()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Bot user: %+v\n", botUser)
	return &TelegramApi{
		bot:     bot,
		todoBot: todoBot,
	}
}

func (t *TelegramApi) Run() error {
	updates, _ := t.bot.UpdatesViaLongPolling(nil)
	defer t.bot.StopLongPolling()

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			userState, err := t.todoBot.GetUserState(chatID)
			if err != nil {
				fmt.Println(err)
				continue
			}

			switch userState {
			case state.Default:
				if update.Message.Text == "/newTask" {
					_, err = t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							"Send task name",
						),
					)
					if err != nil {
						fmt.Println(err)
						continue
					}
					err = t.todoBot.SetUserState(chatID, state.WaitingForNewTaskName)
					if err != nil {
						fmt.Println(err)
						continue
					}
				} else if update.Message.Text == "/deleteTask" {
					_, err = t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							"Send task name",
						),
					)
					if err != nil {
						fmt.Println(err)
						continue
					}
					err = t.todoBot.SetUserState(chatID, state.WaitingForTaskNameToBeDeleted)
					if err != nil {
						fmt.Println(err)
						continue
					}
				} else if update.Message.Text == "/listOfTask" {
					message, err := t.todoBot.GetListOfTasks()
					if err != nil {
						fmt.Println(err)
						continue
					}
					_, err = t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							message,
						),
					)
				} else {
					_, err = t.bot.SendMessage(tu.Messagef(
						tu.ID(update.Message.Chat.ID),
						"Hello %s!", update.Message.From.FirstName,
					))
				}
			case state.WaitingForTaskNameToBeDeleted:
				err = t.todoBot.DeleteTask(update.Message.Text)
				if err != nil {
					fmt.Println(err)
					continue
				}
				_, err = t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Task deleted successfully ",
					),
				)
				err = t.todoBot.SetUserState(chatID, state.Default)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case state.WaitingForNewTaskName:
				taskID, err := t.todoBot.CreateNewTask(chatID)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = t.todoBot.SetTaskName(taskID, update.Message.Text)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = t.todoBot.SetUserState(chatID, state.WaitingForNewTaskDescription)
				if err != nil {
					fmt.Println(err)
					continue
				}
				_, err = t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Send task description",
					),
				)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case state.WaitingForNewTaskDescription:
				taskID, err := t.todoBot.GetTaskIDInCreationStatus(chatID)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = t.todoBot.SetTaskDescription(taskID, update.Message.Text)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = t.todoBot.SetUserState(chatID, state.Default)
				if err != nil {
					fmt.Println(err)
					continue
				}
				_, err = t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Task created",
					),
				)
			}
		}
	}
	return nil
}
