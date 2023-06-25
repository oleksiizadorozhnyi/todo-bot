package telegram

import (
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"os"
	"telegramBot/pkg/adapter/cache/redis"
	todoBot "telegramBot/pkg/adapter/todobot"
	"telegramBot/pkg/model/state/telegram"
	"telegramBot/pkg/model/state/user"
)

type Telegram struct {
	bot     *telego.Bot
	todoBot *todoBot.TodoBot
	cache   *redis.Cache
}

func New(todoBot *todoBot.TodoBot, token string, cache *redis.Cache) *Telegram {
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
	return &Telegram{
		bot:     bot,
		todoBot: todoBot,
		cache:   cache,
	}
}

func (t *Telegram) Run(ctx context.Context) error {
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
			err := t.cache.Set(ctx, chatID, update.Message.MessageID)
			if err != nil {
				zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
				continue
			}
		} else {
			continue
		}

		button := false
		if len(action) > 17 {
			if action[:17] == "/buttonDeleteTask" {
				err := t.todoBot.SetUserState(chatID, user.WaitingForTaskNameToBeDeleted)
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
		case user.Default:
			switch action {
			case telegram.StartState:
				err := t.startHandler(ctx, chatID, update.Message.Chat.FirstName)
				if err != nil {
					zap.L().Error("Run() -> t.startHandler()", zap.Error(err))
					continue
				}
			case telegram.NewTaskState:
				err := t.newTaskHandler(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.newTaskHandler()", zap.Error(err))
					continue
				}
			case telegram.DeleteTaskState:
				err := t.deleteTaskHandler(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.deleteTaskHandler()", zap.Error(err))
					continue
				}
			case telegram.ListOfTasksState:
				err = t.listOfTasksHandler(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.listOfTasksHandler()", zap.Error(err))
					continue
				}
			case telegram.CancelLastActionState:
				err = t.cancelLastActionHandler(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.cancelLastAction()", zap.Error(err))
					continue
				}
			default:
				err = t.defaultHandler(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.defaultHandler()", zap.Error(err))
					continue
				}
			}
		case user.WaitingForTaskNameToBeDeleted:
			if action == telegram.NewTaskState ||
				action == telegram.ListOfTasksState || action == telegram.DeleteTaskState || action == telegram.StartState {
				messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID),
					"Finish your last action or /cancelLastAction"))
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
					continue
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
				}
				continue
			} else if action == telegram.CancelLastActionState {
				err = t.todoBot.SetUserState(chatID, user.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				err = t.deleteMessages(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.deleteMessages()", zap.Error(err))
					continue
				}

				messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
					continue
				}
				continue
			}
			err = t.deleteMessages(ctx, chatID)
			if err != nil {
				zap.L().Error("Run() -> t.deleteMessages()", zap.Error(err))
				continue
			}

			message, err := t.todoBot.DeleteTask(action)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.DeleteTask()", zap.Error(err))
				continue
			}
			if button {
				err = t.getListOfTasks(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.getListOfTasks()", zap.Error(err))
					continue
				}
			}

			messageInfo, err := t.bot.SendMessage(
				tu.Message(tu.ID(chatID), message))
			if err != nil {
				zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
				continue
			}
			err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
			if err != nil {
				zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
				continue
			}

			err = t.todoBot.SetUserState(chatID, user.Default)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
				continue
			}
			err = t.menu(ctx, chatID)
			if err != nil {
				zap.L().Error("Run() -> t.menu()", zap.Error(err))
				continue
			}
		case user.WaitingForNewTaskName:
			if action == telegram.CancelLastActionState {
				err = t.todoBot.SetUserState(chatID, user.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				err = t.todoBot.DeleteNotFinishedTask(chatID)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
					continue
				}
				err = t.deleteMessages(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.deleteMessages()", zap.Error(err))
					continue
				}

				messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
					continue
				}
				err = t.menu(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.menu()", zap.Error(err))
					continue
				}
				continue
			}
			if action == telegram.NewTaskState ||
				action == "/allButtons" || action == telegram.ListOfTasksState ||
				action == telegram.DeleteTaskState || action == telegram.StartState {
				messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Finish your last action or /cancelLastAction"))
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
					continue
				}
				continue
			}
			taskID, err := t.todoBot.GetTaskIDInCreationStatus(chatID)
			err = t.todoBot.SetTaskName(taskID, action)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.SetTaskName()", zap.Error(err))
				continue
			}

			err = t.todoBot.SetUserState(chatID, user.WaitingForNewTaskDescription)
			if err != nil {
				zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
				continue
			}

			inlineKeyboard := tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton("Cancel task creation").
						WithCallbackData(telegram.CancelLastActionState),
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
			err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
			if err != nil {
				zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
				continue
			}
		case user.WaitingForNewTaskDescription:
			if action == telegram.CancelLastActionState {
				err = t.todoBot.SetUserState(chatID, user.Default)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}

				err = t.todoBot.DeleteNotFinishedTask(chatID)
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.DeleteNotFinishedTask()", zap.Error(err))
					continue
				}
				err = t.deleteMessages(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.deleteMessages()", zap.Error(err))
					continue
				}

				messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "Last action canceled"))
				if err != nil {
					zap.L().Error("Run() -> t.todoBot.SetUserState()", zap.Error(err))
					continue
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
					continue
				}
				err = t.menu(ctx, chatID)
				if err != nil {
					zap.L().Error("Run() -> t.menu()", zap.Error(err))
					continue
				}
				continue
			}
			if action == telegram.NewTaskState || action == telegram.ListOfTasksState ||
				action == telegram.DeleteTaskState || action == telegram.StartState {
				messageInfo, err := t.bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Finish your last action or /cancelLastAction",
					),
				)
				if err != nil {
					zap.L().Error("Run() -> t.bot.SendMessage()", zap.Error(err))
				}
				err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
				if err != nil {
					zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
					continue
				}
				continue
			}
			err = t.deleteMessages(ctx, chatID)
			if err != nil {
				zap.L().Error("Run() -> t.deleteMessages()", zap.Error(err))
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

			err = t.todoBot.SetUserState(chatID, user.Default)
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
			err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
			if err != nil {
				zap.L().Error("Run() -> t.cache.Set()", zap.Error(err))
				continue
			}
			err = t.menu(ctx, chatID)
			if err != nil {
				zap.L().Error("Run() -> t.menu()", zap.Error(err))
				continue
			}
		}
	}
	return nil
}

func (t *Telegram) deleteMessages(ctx context.Context, chatID int64) error {
	messageIDs, err := t.cache.Get(ctx, chatID)
	if err != nil {
		return err
	}
	for _, v := range messageIDs {
		err = t.bot.DeleteMessage(&telego.DeleteMessageParams{ChatID: tu.ID(chatID), MessageID: v})
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Telegram) getListOfTasks(ctx context.Context, chatID int64) error {
	tasks, err := t.todoBot.GetListOfTasks(chatID)
	if err != nil {
		return err
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
			return err
		}
		err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Telegram) menu(ctx context.Context, chatID int64) error {
	keyboard := tu.Keyboard(
		tu.KeyboardRow(
			tu.KeyboardButton(telegram.NewTaskState),
			tu.KeyboardButton(telegram.ListOfTasksState),
		),
		tu.KeyboardRow(
			tu.KeyboardButton(telegram.DeleteTaskState),
		),
	).WithResizeKeyboard().WithInputFieldPlaceholder("Select something").
		WithOneTimeKeyboard()
	message := tu.Message(
		tu.ID(chatID),
		"Menu: ",
	).WithReplyMarkup(keyboard)
	messageInfo, err := t.bot.SendMessage(message)
	if err != nil {
		return err
	}
	err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
	if err != nil {
		return err
	}
	return nil
}
