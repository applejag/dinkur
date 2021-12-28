package dinkurdb

import (
	"time"
)

type CommonFields struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	task_Field_End = "End"

	task_Column_End = "end"
)

type Task struct {
	CommonFields
	Name  string    `gorm:"not null;default:''"`
	Start time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	End   *time.Time
}

type Migration struct {
	CommonFields
	Version int
}
