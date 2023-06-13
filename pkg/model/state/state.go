package state

const (
	Default = iota
	WaitingForNewTaskName
	WaitingForNewTaskDescription
	WaitingForTaskNameToBeDeleted
)
