package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"telegramBot/pkg/model/task"
)

type Storage struct {
	database *sql.DB
}

func New() *Storage {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        userID INTEGER,
        taskName TEXT,
        taskDescription TEXT,
        taskStatus INTEGER
    )`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        userID INTEGER,
        state INTEGER
    )`)
	if err != nil {
		log.Fatal(err)
	}
	return &Storage{
		database: db,
	}
}

func (s *Storage) SetUserState(userID int64, state int) error {
	var count int
	err := s.database.QueryRow("SELECT COUNT(*) FROM users WHERE userID = ?", userID).Scan(&count)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage.go -> SetUserState() -> s.database.QueryRow() %s", err.Error()))
	}

	if count > 0 {
		_, err = s.database.Exec("UPDATE users SET state = ? WHERE userID = ?", state, userID)
		if err != nil {
			return errors.New(fmt.Sprintf("Storage.go -> SetUserState() -> s.database.Exec() %s", err.Error()))
		}
	} else {
		_, err := s.database.Exec("INSERT INTO users (userID, state) VALUES (?, ?)", userID, state)
		if err != nil {
			return errors.New(fmt.Sprintf("Storage.go -> SetUserState() -> s.database.Exec() %s", err.Error()))
		}
	}
	return nil
}

func (s *Storage) GetUserState(userID int64) (int, error) {
	rows, err := s.database.Query("SELECT state FROM users WHERE userID = ?", userID)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Storage.go -> GetUserState() -> s.database.Query() %s", err.Error()))
	}
	defer rows.Close()
	var state int
	for rows.Next() {
		err := rows.Scan(&state)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("Storage.go -> GetUserState() -> rows.Scan() %s", err.Error()))
		}
	}
	return state, nil
}

func (s *Storage) CreateNewTask(userID int64) (taskID int64, err error) {
	result, err := s.database.Exec("INSERT INTO tasks (userID) VALUES (?)", userID)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Storage.go -> SaveTask() -> s.database.Exec() %s", err.Error()))
	}

	taskID, err = result.LastInsertId()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Storage.go -> CreateNewTask() -> result.LastInsertId() %s", err.Error()))
	}
	return taskID, nil
}

func (s *Storage) SetTaskName(taskID int64, taskName string) (err error) {
	_, err = s.database.Exec("UPDATE tasks SET taskName = ? WHERE id = ?", taskName, taskID)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage.go -> SetTaskName() -> s.database.Exec() %s", err.Error()))
	}
	return nil
}

func (s *Storage) GetTaskName(taskID int64) (string, error) {
	rows, err := s.database.Query("SELECT taskName FROM tasks WHERE taskID = ?", taskID)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Storage.go -> GetTaskName() -> s.database.Query() %s", err.Error()))
	}
	defer rows.Close()
	var taskName string
	for rows.Next() {
		err := rows.Scan(&taskName)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Storage.go -> GetTaskName() -> rows.Scan() %s", err.Error()))
		}
	}
	return taskName, nil
}

func (s *Storage) SetTaskDescription(taskID int64, taskDescription string) error {
	_, err := s.database.Exec("UPDATE tasks SET taskDescription = ? WHERE id = ?", taskDescription, taskID)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage.go -> SetTaskDescription() -> s.database.Exec() %s", err.Error()))
	}
	return nil
}

func (s *Storage) GetTaskDescription(taskID int64) (string, error) {
	rows, err := s.database.Query("SELECT taskDescription FROM tasks WHERE taskID = ?", taskID)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Storage.go -> GetTaskDescription() -> s.database.Query() %s",
			err.Error()))
	}
	defer rows.Close()
	var taskDescription string
	for rows.Next() {
		err := rows.Scan(&taskDescription)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Storage.go -> GetTaskDescription() -> rows.Scan() %s", err.Error()))
		}
	}
	return taskDescription, nil
}

func (s *Storage) SetTaskStatus(taskID int64, taskStatus int) error {
	_, err := s.database.Exec("UPDATE tasks SET taskStatus = ? WHERE id = ?", taskStatus, taskID)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage.go -> SetTaskStatus() -> s.database.Exec() %s", err.Error()))
	}
	return nil
}

func (s *Storage) GetTaskIDInCreationStatus(userID int64) (int64, error) {
	rows, err := s.database.Query("SELECT id FROM tasks WHERE userID = ?", userID)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Storage.go -> GetTaskIDInCreationStatus() -> s.database.Query() %s",
			err.Error()))
	}
	defer rows.Close()
	var taskID int64
	for rows.Next() {
		err := rows.Scan(&taskID)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("Storage.go -> GetTaskIDInCreationStatus() -> rows.Scan() %s",
				err.Error()))
		}
	}
	return taskID, nil
}

func (s *Storage) DeleteTask(taskName string) error {
	_, err := s.database.Exec("DELETE FROM tasks WHERE taskName = ?", taskName)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage.go -> DeleteTask() -> s.database.Exec() %s",
			err.Error()))
	}
	return nil
}

func (s *Storage) GetListOfTasks() (string, error) {
	var tasks []string

	rows, err := s.database.Query("SELECT taskName, taskDescription FROM tasks")
	if err != nil {
		return "", errors.New(fmt.Sprintf("GetListOfTasks() -> s.database.Query() %s", err.Error()))
	}
	defer rows.Close()

	for rows.Next() {
		var taskold task.Task
		if err := rows.Scan(&taskold.TaskName, &taskold.TaskDescription); err != nil {
			return "", errors.New(fmt.Sprintf("GetListOfTasks() -> rows.Scan() %s", err.Error()))
		}
		tasks = append(tasks, "\n\nTask:\n{", taskold.TaskName, ": ", taskold.TaskDescription, " }\n\n")
	}
	return fmt.Sprintf("List of your task: %v", tasks), nil
}
