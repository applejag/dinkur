package dinkurdb

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func nilNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (c *client) ActiveTask() (*Task, error) {
	if c.db == nil {
		return nil, ErrNotConnected
	}
	return getActiveTask(c.db)
}

func getActiveTask(db *gorm.DB) (*Task, error) {
	var task Task
	err := db.Where(Task{End: nil}, task_Field_End).First(&task).Error
	if err != nil {
		return nil, nilNotFoundError(err)
	}
	return &task, nil
}

type NewTask struct {
	Name  string
	Start *time.Time
	End   *time.Time
}

type StartedTask struct {
	New      Task
	Previous *Task
}

func (c *client) StartTask(task NewTask) (StartedTask, error) {
	if c.db == nil {
		return StartedTask{}, ErrNotConnected
	}
	if task.Name == "" {
		return StartedTask{}, ErrTaskNameEmpty
	}
	var start time.Time
	if task.Start != nil {
		start = *task.Start
	} else {
		start = time.Now()
	}
	if task.End != nil && task.End.Before(start) {
		return StartedTask{}, ErrTaskEndBeforeStart
	}
	dbTask := Task{
		Name:  task.Name,
		Start: start,
		End:   task.End,
	}
	return startTask(c.db, dbTask)
}

func (c *client) StopActiveTask() (bool, error) {
	if c.db == nil {
		return false, ErrNotConnected
	}
	rows, err := stopAllTasks(c.db)
	return rows > 0, err
}

func startTask(db *gorm.DB, task Task) (StartedTask, error) {
	var activeTask *Task
	db.Transaction(func(tx *gorm.DB) error {
		var err error
		activeTask, err = getActiveTask(tx)
		if err != nil {
			return err
		}
		_, err = stopAllTasks(tx)
		if err != nil {
			return err
		}
		err = db.Create(&task).Error
		if err != nil {
			return err
		}
		return nil
	})
	return StartedTask{
		New:      task,
		Previous: activeTask,
	}, nil
}

func stopAllTasks(db *gorm.DB) (int64, error) {
	res := db.Model(&Task{}).
		Where(&Task{End: nil}, task_Field_End).
		Update(task_Column_End, time.Now())
	return res.RowsAffected, res.Error
}
