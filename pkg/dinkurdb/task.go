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
	"fmt"
	"time"
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
