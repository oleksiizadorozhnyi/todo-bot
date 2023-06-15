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
		var action string
		var chatID int64
		if update.CallbackQuery != nil {
			action = update.CallbackQuery.Data
			chatID = update.CallbackQuery.Message.Chat.ID
		} else if update.Message != nil {
			action = update.Message.Text
			chatID = update.Message.Chat.ID
		}

		if action != "" {
			if len(action) > 11 {
				if action[:11] == "/deleteTask" {
					err := t.todoBot.SetUserState(chatID, state.WaitingForTaskNameToBeDeleted)
					if err != nil {
						zap.L().Error("Run() -> /newTask -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					action = action[11:]
				}
			}

			userState, err := t.todoBot.GetUserState(chatID)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.GetUserState()", zap.Error(err))
				continue
			}

			//if chatID == 287757469 {
			//	_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Nikisha sosi pisun"))
			//	if err != nil {
			//		zap.L().Error("Run() -> /deleteTask -> t.bot.SendMessage()", zap.Error(err))
			//		continue
			//	}
			//	continue
			//}
			switch userState {
			case state.Default:
				switch action {
				case "/newTask":
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
					inlineKeyboard := tu.InlineKeyboard(
						tu.InlineKeyboardRow(
							tu.InlineKeyboardButton("Cancel task creation").
								WithCallbackData("/cancelLastAction"),
						),
					)
					message := tu.Message(
						tu.ID(chatID),
						"Send task name",
					).WithReplyMarkup(inlineKeyboard)

					_, _ = t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				case "/deleteTask":
					err = t.todoBot.SetUserState(chatID, state.WaitingForTaskNameToBeDeleted)
					if err != nil {
						zap.L().Error("Run() -> /deleteTask -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					inlineKeyboard := tu.InlineKeyboard(
						tu.InlineKeyboardRow(
							tu.InlineKeyboardButton("Cancel task deletion").
								WithCallbackData("/cancelLastAction"),
						),
					)
					message := tu.Message(
						tu.ID(chatID),
						"Send task name",
					).WithReplyMarkup(inlineKeyboard)

					_, _ = t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				case "/listOfTasks":
					tasks, err := t.todoBot.GetListOfTasks(chatID)
					if err != nil {
						zap.L().Error("Run() -> /listOfTasks -> t.todoBot.GetListOfTasks()", zap.Error(err))
						continue
					}
					for _, task := range tasks {
						inlineKeyboard := tu.InlineKeyboard(
							tu.InlineKeyboardRow(
								tu.InlineKeyboardButton("Delete this task").
									WithCallbackData("/deleteTask" + task.TaskName),
							),
						)
						message := tu.Messagef(
							tu.ID(chatID),
							`Task name:             %s
Task description:   %s`, task.TaskName, task.TaskDescription,
						).WithReplyMarkup(inlineKeyboard)

						_, _ = t.bot.SendMessage(message)
						if err != nil {
							zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
							continue
						}
					}
				case "/allButtons":
					inlineKeyboard := tu.InlineKeyboard(
						tu.InlineKeyboardRow(
							tu.InlineKeyboardButton("Make new task").
								WithCallbackData("/newTask"),
							tu.InlineKeyboardButton("All my tasks").
								WithCallbackData("/listOfTasks"),
						),
						tu.InlineKeyboardRow(
							tu.InlineKeyboardButton("Delete task").
								WithCallbackData("/deleteTask"),
							tu.InlineKeyboardButton("All my tasks").
								WithCallbackData("/listOfTasks"),
						),
					)
					message := tu.Message(
						tu.ID(chatID),
						"All I can do now:",
					).WithReplyMarkup(inlineKeyboard)

					_, _ = t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
				default:

					inlineKeyboard := tu.InlineKeyboard(
						tu.InlineKeyboardRow( // Row 1
							tu.InlineKeyboardButton("See all commands").
								WithCallbackData("/allButtons"),
						),
					)

					message := tu.Messagef(
						tu.ID(chatID),
						"Hello %s! Use buttons below", update.Message.From.FirstName,
					).WithReplyMarkup(inlineKeyboard)

					_, _ = t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}

				}
			case state.WaitingForTaskNameToBeDeleted:
				if action == "/newTask" || action == "/allButtons" ||
					action == "/listOfTasks" || action == "/deleteTask" {
					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID),
						"Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					continue
				} else if action == "/cancelLastAction" {
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
				message, err := t.todoBot.DeleteTask(action)
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
				if action == "/cancelLastAction" {
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
				if action == "/newTask" ||
					action == "/allButtons" || action == "/listOfTasks" ||
					action == "/deleteTask" {
					_, err = t.bot.SendMessage(tu.Message(tu.ID(chatID), "Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					continue
				}
				taskID, err := t.todoBot.GetTaskIDInCreationStatus(chatID)
				err = t.todoBot.SetTaskName(taskID, action)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetTaskName()", zap.Error(err))
					continue
				}

				err = t.todoBot.SetUserState(chatID, state.WaitingForNewTaskDescription)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

				inlineKeyboard := tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("Cancel task creation").
							WithCallbackData("/cancelLastAction"),
					),
				)
				message := tu.Message(
					tu.ID(chatID),
					"Send task description",
				).WithReplyMarkup(inlineKeyboard)

				_, _ = t.bot.SendMessage(message)
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}

			case state.WaitingForNewTaskDescription:
				if action == "/cancelLastAction" {
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
				if action == "/newTask" || action == "/allButtons" || action == "/listOfTasks" ||
					action == "/deleteTask" {
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

				err = t.todoBot.SetTaskDescription(taskID, action)
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
