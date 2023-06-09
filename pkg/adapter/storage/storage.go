package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"telegramBot/pkg/model"
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
        name TEXT,
        description TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}
	return &Storage{
		database: db,
	}
}

func (s *Storage) SaveTask() string {
	name := "Новая задача"
	description := "Описание новой задачи"

	result, err := s.database.Exec("INSERT INTO tasks (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		log.Fatal(err)
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("link saved. id: %d", taskID)
}

func (s *Storage) GetListOfTasks() string {
	var tasks []model.Task

	rows, err := s.database.Query("SELECT id, name FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.Name); err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("Tasks: %v", tasks)
}

func (s *Storage) DeleteTask() string {
	taskID := 6

	stmt, err := s.database.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskID)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("Задача с ID %d удалена", taskID)
}
