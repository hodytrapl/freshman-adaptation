package models

import (
	"time"
)

/*валидация, саша , используй это*/

// Group соответствует таблице groups
type Group struct {
	ID   int64  `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"not null"`
}

// Student соответствует таблице students
type Student struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	FirstName string    `json:"first_name" gorm:"column:first_name;not null"`
	LastName  string    `json:"last_name" gorm:"column:last_name;not null"`
	GroupID   int64     `json:"group_id" gorm:"column:group_id;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TaskStatus определяет возможные статусы задачи
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// Task соответствует таблице tasks
type Task struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	StudentID int64      `json:"student_id" gorm:"column:student_id;not null;index"`
	Title     string     `json:"title" gorm:"not null"`
	Status    TaskStatus `json:"status" gorm:"column:status;not null;default:todo"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}
