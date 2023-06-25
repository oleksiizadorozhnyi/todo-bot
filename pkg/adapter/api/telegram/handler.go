package telegram

import (
	"context"
	tu "github.com/mymmrac/telego/telegoutil"
	"telegramBot/pkg/model/state/telegram"
	"telegramBot/pkg/model/state/user"
)

func (t *Telegram) startHandler(ctx context.Context, chatID int64, firstName string) error {
	err := t.deleteMessages(ctx, chatID)
	if err != nil {
		return err
	}
	_, err = t.bot.SendMessage(tu.Messagef(tu.ID(chatID), "Hello %s!", firstName))
	if err != nil {
		return err
	}
	err = t.menu(ctx, chatID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) newTaskHandler(ctx context.Context, chatID int64) error {
	err := t.deleteMessages(ctx, chatID)
	if err != nil {
		return err
	}
	err = t.todoBot.SetUserState(chatID, user.WaitingForNewTaskName)
	if err != nil {
		return err
	}

	_, err = t.todoBot.CreateNewTask(chatID)
	if err != nil {
		return err
	}
	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Cancel task creation").
				WithCallbackData(telegram.CancelLastActionState),
		),
	)
	message := tu.Message(
		tu.ID(chatID),
		"Send task name",
	).WithReplyMarkup(inlineKeyboard)

	messageInfo, _ := t.bot.SendMessage(message)
	if err != nil {
		return err
	}
	err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) deleteTaskHandler(ctx context.Context, chatID int64) error {
	err := t.deleteMessages(ctx, chatID)
	if err != nil {
		return err
	}
	err = t.todoBot.SetUserState(chatID, user.WaitingForTaskNameToBeDeleted)
	if err != nil {
		return err
	}
	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Cancel task deletion").
				WithCallbackData(telegram.CancelLastActionState),
		),
	)
	message := tu.Message(
		tu.ID(chatID),
		"Send task name",
	).WithReplyMarkup(inlineKeyboard)

	messageInfo, _ := t.bot.SendMessage(message)
	if err != nil {
		return err
	}
	err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) listOfTasksHandler(ctx context.Context, chatID int64) error {
	err := t.deleteMessages(ctx, chatID)
	if err != nil {
		return err
	}
	err = t.getListOfTasks(ctx, chatID)
	if err != nil {
		return err
	}
	err = t.menu(ctx, chatID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) cancelLastActionHandler(ctx context.Context, chatID int64) error {
	messageInfo, err := t.bot.SendMessage(tu.Message(tu.ID(chatID), "There is nothing to cancel"))
	if err != nil {
		return err
	}
	err = t.cache.Set(ctx, chatID, messageInfo.MessageID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) defaultHandler(ctx context.Context, chatID int64) error {
	err := t.deleteMessages(ctx, chatID)
	if err != nil {
		return err
	}
	err = t.menu(ctx, chatID)
	if err != nil {
		return err
	}
	return nil
}
