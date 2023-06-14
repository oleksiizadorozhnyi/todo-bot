package telegramApi

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
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
		zap.L().Error("New() -> telego.NewBot()", zap.Error(err))
		os.Exit(1)
	}
	botUser, err := bot.GetMe()
	if err != nil {
		zap.L().Error("New() -> bot.GetMe()", zap.Error(err))
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
				zap.L().Error("Run() -> t.todoBot.GetUserState()", zap.Error(err))
				continue
			}

			switch userState {
			case state.Default:
				switch update.Message.Text {
				case "/newTask":
					_, err = t.bot.SendMessage(
						tu.Message(tu.ID(chatID), "Send task name"))
					if err != nil {
						zap.L().Error("Run() -> /newTask -> t.bot.SendMessage()", zap.Error(err))
						continue
					}

					err = t.todoBot.SetUserState(chatID, state.WaitingForNewTaskName)
					if err != nil {
						zap.L().Error("Run() -> /newTask -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}

					_, err := t.todoBot.CreateNewTask(chatID)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.CreateNewTask()", zap.Error(err))
						continue
					}
				case "/deleteTask":
					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Send task name"))
					if err != nil {
						zap.L().Error("Run() -> /deleteTask -> t.bot.SendMessage()", zap.Error(err))
						continue
					}

					err = t.todoBot.SetUserState(chatID, state.WaitingForTaskNameToBeDeleted)
					if err != nil {
						zap.L().Error("Run() -> /deleteTask -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
				case "/listOfTasks":
					message, err := t.todoBot.GetListOfTasks(chatID)
					if err != nil {
						zap.L().Error("Run() -> /listOfTasks -> t.todoBot.GetListOfTasks()", zap.Error(err))
						continue
					}
					_, err = t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							message,
						),
					)
					if err != nil {
						zap.L().Error("Run() -> /listOfTasks -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				case "/help":
					_, err = t.bot.SendMessage(tu.Messagef(
						tu.ID(update.Message.Chat.ID),
						"All commands: /newTask , /deleteTask , /listOfTasks , /help",
					))
					if err != nil {
						zap.L().Error("Run() -> /help -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				default:
					_, err = t.bot.SendMessage(tu.Messagef(
						tu.ID(update.Message.Chat.ID),
						"Hello %s! Use /help", update.Message.From.FirstName,
					))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				}
			case state.WaitingForTaskNameToBeDeleted:
				if update.Message.Text == "/newTask" || update.Message.Text == "/help" ||
					update.Message.Text == "/listOfTasks" || update.Message.Text == "/deleteTask" {
					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID),
						"Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					continue
				} else if update.Message.Text == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}

					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					continue
				}
				message, err := t.todoBot.DeleteTask(update.Message.Text)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.DeleteTask()", zap.Error(err))
					continue
				}

				_, err = t.bot.SendMessage(
					tu.Message(tu.ID(chatID), message))
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

				err = t.todoBot.SetUserState(chatID, state.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

			case state.WaitingForNewTaskName:
				if update.Message.Text == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					err = t.todoBot.DeleteNotFinishedTask()
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
						continue
					}

					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					continue
				}
				if update.Message.Text == "/newTask" ||
					update.Message.Text == "/help" || update.Message.Text == "/listOfTasks" ||
					update.Message.Text == "/deleteTask" {
					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					continue
				}
				taskID, err := t.todoBot.GetTaskIDInCreationStatus(chatID)
				err = t.todoBot.SetTaskName(taskID, update.Message.Text)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetTaskName()", zap.Error(err))
					continue
				}

				err = t.todoBot.SetUserState(chatID, state.WaitingForNewTaskDescription)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

				_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Send task description"))
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}

			case state.WaitingForNewTaskDescription:
				if update.Message.Text == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}

					err = t.todoBot.DeleteNotFinishedTask()
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
						continue
					}

					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					continue
				}
				if update.Message.Text == "/newTask" || update.Message.Text == "/help" || update.Message.Text == "/listOfTasks" ||
					update.Message.Text == "/deleteTask" {
					_, err = t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							"Finish your last action or /cancelLastAction",
						),
					)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					continue
				}
				taskID, err := t.todoBot.GetTaskIDInCreationStatus(chatID)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.GetTaskIDInCreationStatus()", zap.Error(err))
					continue
				}

				err = t.todoBot.SetTaskDescription(taskID, update.Message.Text)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetTaskDescription()", zap.Error(err))
					continue
				}

				err = t.todoBot.SetUserState(chatID, state.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

				_, err = t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Task created",
					),
				)
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}
			}
		}
	}
	return nil
}
