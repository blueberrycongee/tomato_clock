package model

import "time"

const (
	RepeatNone    = "none"
	RepeatDaily   = "daily"
	RepeatWeekly  = "weekly"
	RepeatWorkday = "workday"
)

type Task struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	Note       string     `json:"note,omitempty"`
	IsDone     bool       `json:"is_done"`
	RepeatRule string     `json:"repeat_rule"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Label      string     `json:"label"`
	DueDate    *time.Time `json:"due_date,omitempty"`
}

func CreateTask(t *Task) error {
	return AddTask(t)
}

func ListTasks() ([]Task, error) {
	return AllTasks(), nil
}
