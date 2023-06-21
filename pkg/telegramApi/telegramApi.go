package telegramApi

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"os"
	"telegramBot/pkg/adapter/storage/messageIdStorage"
	todo_bot "telegramBot/pkg/adapter/todo-bot"
	"telegramBot/pkg/model/state"
)

type TelegramApi struct {
	bot              *telego.Bot
	todoBot          *todo_bot.TodoBot
	messageIdStorage *messageIdStorage.MessageIdStorage
}

func New(todoBot *todo_bot.TodoBot, token string, messageIdStorage *messageIdStorage.MessageIdStorage) *TelegramApi {
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
		bot:              bot,
		todoBot:          todoBot,
		messageIdStorage: messageIdStorage,
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
			t.messageIdStorage.Set(chatID, update.Message.MessageID)
		}

		if action != "" {
			button := false
			if len(action) > 17 {
				if action[:17] == "/buttonDeleteTask" {
					err := t.todoBot.SetUserState(chatID, state.WaitingForTaskNameToBeDeleted)
					if err != nil {
						zap.L().Error("Run() -> /newTask -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					action = action[17:]
					button = true
				}
			}

			userState, err := t.todoBot.GetUserState(chatID)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.GetUserState()", zap.Error(err))
				continue
			}

			switch userState {
			case state.Default:
				switch action {
				case "/start":
					t.deleteMessages(chatID)
					_, err := t.bot.SendMessage(tu.Messagef(tu.ID(chatID), "Hello %s!", update.Message.Chat.FirstName))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					t.menu(chatID)
				case "/newTask":
					t.deleteMessages(chatID)
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

					messageInfo, _ := t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
				case "/deleteTask":
					t.deleteMessages(chatID)
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

					messageInfo, _ := t.bot.SendMessage(message)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
				case "/listOfTasks":
					t.deleteMessages(chatID)
					t.getListOfTasks(chatID)
					t.menu(chatID)
				case "/cancelLastAction":
					messageInfo, _ := t.bot.SendMessage(tu.Message(tu.ID(chatID), "There is nothing to cancel"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
				default:
					t.deleteMessages(chatID)
					t.menu(chatID)
				}
			case state.WaitingForTaskNameToBeDeleted:
				if action == "/newTask" || action == "/allButtons" ||
					action == "/listOfTasks" || action == "/deleteTask" || action == "/start" {
					messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID),
						"Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
					continue
				} else if action == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					t.deleteMessages(chatID)

					messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
					continue
				}
				t.deleteMessages(chatID)

				message, err := t.todoBot.DeleteTask(action)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.DeleteTask()", zap.Error(err))
					continue
				}
				if button {
					t.getListOfTasks(chatID)
				}

				messageInfo, err := t.bot.SendMessage(
					tu.Message(tu.ID(chatID), message))
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				t.messageIdStorage.Set(chatID, messageInfo.MessageID)

				err = t.todoBot.SetUserState(chatID, state.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				t.menu(chatID)
			case state.WaitingForNewTaskName:
				if action == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					err = t.todoBot.DeleteNotFinishedTask(chatID)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
						continue
					}
					t.deleteMessages(chatID)

					messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
					t.menu(chatID)
					continue
				}
				if action == "/newTask" ||
					action == "/allButtons" || action == "/listOfTasks" ||
					action == "/deleteTask" || action == "/start" {
					messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Finish your last action or /cancelLastAction"))
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
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

				messageInfo, _ := t.bot.SendMessage(message)
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}
				t.messageIdStorage.Set(chatID, messageInfo.MessageID)

			case state.WaitingForNewTaskDescription:
				if action == "/cancelLastAction" {
					err = t.todoBot.SetUserState(chatID, state.Default)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}

					err = t.todoBot.DeleteNotFinishedTask(chatID)
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
						continue
					}
					t.deleteMessages(chatID)

					messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
					if err != nil {
						zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
						continue
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
					t.menu(chatID)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
						continue
					}
					continue
				}
				if action == "/newTask" || action == "/allButtons" || action == "/listOfTasks" ||
					action == "/deleteTask" || action == "/start" {
					messageInfo, err := t.bot.SendMessage(
						tu.Message(
							tu.ID(chatID),
							"Finish your last action or /cancelLastAction",
						),
					)
					if err != nil {
						zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					}
					t.messageIdStorage.Set(chatID, messageInfo.MessageID)
					continue
				}
				t.deleteMessages(chatID)
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

				messageInfo, err := t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Task created",
					),
				)
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}
				t.messageIdStorage.Set(chatID, messageInfo.MessageID)
				t.menu(chatID)
			}
		}
	}
	return nil
}

func (t *TelegramApi) deleteMessages(chatID int64) {
	messageIDs := t.messageIdStorage.Get(chatID)
	fmt.Println(messageIDs)
	for _, v := range messageIDs {
		err := t.bot.DeleteMessage(&telego.DeleteMessageParams{ChatID: tu.ID(chatID), MessageID: v})
		if err != nil {
			fmt.Println(err)
			zap.L().Error("deleteMessages()", zap.Error(err))
		}
	}
}

func (t *TelegramApi) getListOfTasks(chatID int64) {
	tasks, err := t.todoBot.GetListOfTasks(chatID)
	if err != nil {
		zap.L().Error("Run() -> /listOfTasks -> t.todoBot.GetListOfTasks()", zap.Error(err))
		return
	}
	for _, task := range tasks {
		inlineKeyboard := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("Delete this task").
					WithCallbackData("/buttonDeleteTask" + task.TaskName),
			),
		)
		message := tu.Messagef(
			tu.ID(chatID),
			`Task name:             %s
Task description:   %s`, task.TaskName, task.TaskDescription,
		).WithReplyMarkup(inlineKeyboard)

		messageInfo, _ := t.bot.SendMessage(message)
		if err != nil {
			zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
			return
		}
		t.messageIdStorage.Set(chatID, messageInfo.MessageID)
	}
	return
}

func (t *TelegramApi) menu(chatID int64) {
	keyboard := tu.Keyboard(
		tu.KeyboardRow(
			tu.KeyboardButton("/newTask"),
			tu.KeyboardButton("/listOfTasks"),
		),
		tu.KeyboardRow(
			tu.KeyboardButton("/deleteTask"),
		),
	).WithResizeKeyboard().WithInputFieldPlaceholder("Select something").
		WithOneTimeKeyboard()
	message := tu.Message(
		tu.ID(chatID),
		"Menu: ",
	).WithReplyMarkup(keyboard)
	messageInfo, err := t.bot.SendMessage(message)
	if err != nil {
		zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
		return
	}
	t.messageIdStorage.Set(chatID, messageInfo.MessageID)
}
