// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package dinkurdb

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dinkur/dinkur/pkg/timeutil"
)

func (c *client) ActiveTask() (*Task, error) {
	if c.db == nil {
		return nil, ErrNotConnected
	}
	var task Task
	err := c.db.Where(Task{End: nil}, task_Field_End).First(&task).Error
	if err != nil {
		return nil, nilNotFoundError(err)
	}
	return &task, nil
}

func (c *client) GetTask(id uint) (Task, error) {
	if c.db == nil {
		return Task{}, ErrNotConnected
	}
	var task Task
	err := c.db.First(&task, id).Error
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

type SearchTask struct {
	Start *time.Time
	End   *time.Time
	Limit uint

	Shorthand timeutil.TimeSpanShorthand
}

var (
	task_SQL_End_LE_and_not_null = fmt.Sprintf("(%[1]s <= ? AND %[1]s IS NOT NULL)", task_Column_End)
	task_SQL_End_LE_or_null      = fmt.Sprintf("(%[1]s <= ? OR %[1]s IS NULL)", task_Column_End)
)

func (c *client) ListTasks(search SearchTask) ([]Task, error) {
	if search.Shorthand != timeutil.TimeSpanNone {
		span := search.Shorthand.Span(time.Now())
		if search.Start == nil {
			search.Start = &span.Start
		}
		if search.End == nil {
			search.End = &span.End
		}
	}
	if search.Limit > math.MaxInt {
		return nil, ErrLimitTooLarge
	}
	var tasks []Task
	q := c.db.Model(&Task{}).
		Order(task_Column_Start + " desc").
		Limit(int(search.Limit))
	if search.Start != nil {
		q = q.Where(task_Column_Start+" >= ?", *search.Start)
	}
	if search.End != nil {
		// treat task.End==nil as task.End==time.Now()
		if search.End.Before(time.Now()) {
			// exclude task.End==nil, as end has not passed time.Now() yet
			q = q.Where(task_SQL_End_LE_and_not_null, *search.End)
		} else {
			// include task.End==nil, as end has passed time.Now()
			q = q.Where(task_SQL_End_LE_or_null, *search.End)
		}
	}
	if err := q.Find(&tasks).Error; err != nil {
		return nil, err
	}
	// we sorted in descending order to get the last tasks.
	// fix this by reversing "again"
	reverseTaskSlice(tasks)
	return tasks, nil
}

type EditTask struct {
	ID    *uint
	Name  *string
	Start *time.Time
	End   *time.Time
}

type UpdatedTask struct {
	Old     Task
	Updated Task
}

func (c *client) EditTask(edit EditTask) (UpdatedTask, error) {
	if c.db == nil {
		return UpdatedTask{}, ErrNotConnected
	}
	if edit.Name != nil && *edit.Name == "" {
		return UpdatedTask{}, ErrTaskNameEmpty
	}
	if edit.Start != nil && edit.End != nil && edit.Start.After(*edit.End) {
		return UpdatedTask{}, ErrTaskEndBeforeStart
	}
	var update UpdatedTask
	err := c.transaction(func(tx *client) error {
		task, err := tx.getTaskToEdit(edit.ID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("no task to edit, failed finding latest task: %w", err)
			}
			return fmt.Errorf("get task to edit: %w", err)
		}
		var anyEdit bool
		update.Old = task
		if edit.Name != nil {
			task.Name = *edit.Name
			anyEdit = true
		}
		if edit.Start != nil {
			task.Start = *edit.Start
			anyEdit = true
		}
		if edit.End != nil {
			task.End = edit.End
			anyEdit = true
		}
		if task.Elapsed() < 0 {
			return ErrTaskEndBeforeStart
		}
		if anyEdit {
			if err := tx.db.Save(&task).Error; err != nil {
				return fmt.Errorf("save updated task: %w", err)
			}
		}
		update.Updated = task
		return nil
	})
	return update, err
}

func (c *client) getTaskToEdit(id *uint) (Task, error) {
	var task Task
	err := c.transaction(func(tx *client) error {
		if id != nil {
			taskByID, err := tx.GetTask(*id)
			if err != nil {
				return fmt.Errorf("get task by ID: %d: %w", *id, err)
			}
			task = taskByID
			return nil
		}
		activeTask, err := tx.ActiveTask()
		if err != nil {
			return fmt.Errorf("get active task: %w", err)
		}
		if activeTask != nil {
			task = *activeTask
			return nil
		}
		now := time.Now()
		tasks, err := tx.ListTasks(SearchTask{
			Limit: 1,
			End:   &now,
		})
		if err != nil {
			return fmt.Errorf("list latest 1 task: %w", err)
		}
		if len(tasks) == 0 {
			return ErrNotFound
		}
		task = tasks[0]
		return nil
	})
	return task, err
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
	newTask := Task{
		Name:  task.Name,
		Start: start,
		End:   task.End,
	}
	var activeTask *Task
	c.transaction(func(tx *client) error {
		var err error
		activeTask, err = tx.StopActiveTask()
		if err != nil {
			return err
		}
		err = tx.db.Create(&newTask).Error
		if err != nil {
			return fmt.Errorf("create new active task: %w", err)
		}
		return nil
	})
	return StartedTask{
		New:      newTask,
		Previous: activeTask,
	}, nil
}

func (c *client) StopActiveTask() (*Task, error) {
	if c.db == nil {
		return nil, ErrNotConnected
	}
	var activeTask *Task
	err := c.transaction(func(tx *client) error {
		var err error
		activeTask, err = tx.ActiveTask()
		if err != nil {
			return fmt.Errorf("get previously active task: %w", err)
		}
		_, err = tx.stopAllTasks()
		if err != nil {
			return fmt.Errorf("stop previously active task: %w", err)
		}
		if activeTask != nil {
			updatedTask, err := tx.GetTask(activeTask.ID)
			if err != nil {
				return fmt.Errorf("get updated previously active task: %w", err)
			}
			activeTask = &updatedTask
		}
		return nil
	})
	return activeTask, err
}

func (c *client) stopAllTasks() (bool, error) {
	if c.db == nil {
		return false, ErrNotConnected
	}
	res := c.db.Model(&Task{}).
		Where(&Task{End: nil}, task_Field_End).
		Update(task_Column_End, time.Now())
	return res.RowsAffected > 0, res.Error
}

func reverseTaskSlice(slice []Task) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
