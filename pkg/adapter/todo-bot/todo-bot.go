package todo_bot

import (
	"telegramBot/pkg/model/task"
	"telegramBot/pkg/model/task/status"
)

type Storage interface {
	SetUserState(userID int64, state int) error
	GetUserState(userID int64) (int, error)
	CreateNewTask(userID int64) (taskID int64, err error)
	SetTaskName(taskID int64, taskName string) (err error)
	GetTaskName(taskID int64) (string, error)
	SetTaskDescription(taskID int64, taskDescription string) error
	GetTaskDescription(taskID int64) (string, error)
	SetTaskStatus(taskID int64, taskStatus int) error
	GetTaskIDInCreationStatus(userID int64) (int64, error)
	DeleteTask(taskName string) (string, error)
	DeleteNotFinishedTask(chatId int64) error
	GetListOfTasks(userID int64) ([]task.Task, error)
}

type TodoBot struct {
	storage Storage
}

func New(database Storage) *TodoBot {
	return &TodoBot{
		storage: database,
	}
}

func (s *TodoBot) SetUserState(userID int64, state int) error {
	return s.storage.SetUserState(userID, state)
}

func (s *TodoBot) GetUserState(userID int64) (int, error) {
	return s.storage.GetUserState(userID)
}

func (s *TodoBot) CreateNewTask(userID int64) (taskID int64, err error) {
	taskID, err = s.storage.CreateNewTask(userID)
	if err != nil {
		return 0, err
	}
	err = s.storage.SetTaskStatus(taskID, status.Creating)
	if err != nil {
		return 0, err
	}
	return taskID, nil
}

func (s *TodoBot) SetTaskName(taskID int64, taskName string) (err error) {
	return s.storage.SetTaskName(taskID, taskName)
}

func (s *TodoBot) GetTaskName(taskID int64) (string, error) {
	return s.storage.GetTaskName(taskID)
}

func (s *TodoBot) SetTaskDescription(taskID int64, taskDescription string) error {
	err := s.storage.SetTaskDescription(taskID, taskDescription)
	if err != nil {
		return err
	}
	err = s.storage.SetTaskStatus(taskID, status.Created)
	if err != nil {
		return err
	}
	return nil
}

func (s *TodoBot) GetTaskDescription(taskID int64) (string, error) {
	return s.storage.GetTaskDescription(taskID)
}

func (s *TodoBot) GetTaskIDInCreationStatus(userID int64) (int64, error) {
	return s.storage.GetTaskIDInCreationStatus(userID)
}

func (s *TodoBot) DeleteTask(taskName string) (string, error) {
	return s.storage.DeleteTask(taskName)
}

func (s *TodoBot) DeleteNotFinishedTask(chatId int64) error {
	return s.storage.DeleteNotFinishedTask(chatId)
}

func (s *TodoBot) GetListOfTasks(userID int64) ([]task.Task, error) {
	return s.storage.GetListOfTasks(userID)
}

func (s *TodoBot) SetTaskStatus(taskID int64, taskStatus int) error {
	return s.storage.SetTaskStatus(taskID, taskStatus)
}
