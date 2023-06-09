package todo_bot

type Storage interface {
	SaveTask() string
	GetListOfTasks() string
	DeleteTask() string
}

type TodoBot struct {
	storage Storage
}

func New(database Storage) *TodoBot {
	return &TodoBot{
		storage: database,
	}
}

func (t *TodoBot) NewTask() string {
	return t.storage.SaveTask()
}

func (t *TodoBot) ListOfTasks() string {
	return t.storage.GetListOfTasks()
}

func (t *TodoBot) DeleteTask() string {
	return t.storage.DeleteTask()
}
