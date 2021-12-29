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

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
)

func (c *client) ActiveTask() (*dinkur.Task, error) {
	dbTask, err := c.activeDBTask()
	if err != nil {
		return nil, err
	}
	return convTaskPtr(dbTask), nil
}

func (c *client) activeDBTask() (*Task, error) {
	if c.db == nil {
		return nil, ErrNotConnected
	}
	var dbTask Task
	err := c.db.Where(Task{End: nil}, task_Field_End).First(&dbTask).Error
	if err != nil {
		return nil, nilNotFoundError(err)
	}
	return &dbTask, nil
}

func (c *client) GetTask(id uint) (dinkur.Task, error) {
	dbTask, err := c.getDBTask(id)
	if err != nil {
		return dinkur.Task{}, err
	}
	return convTask(dbTask), nil
}

func (c *client) getDBTask(id uint) (Task, error) {
	if c.db == nil {
		return Task{}, ErrNotConnected
	}
	var dbTask Task
	err := c.db.First(&dbTask, id).Error
	if err != nil {
		return Task{}, err
	}
	return dbTask, nil
}

var (
	task_SQL_End_LE_and_not_null = fmt.Sprintf("(%[1]s <= ? AND %[1]s IS NOT NULL)", task_Column_End)
	task_SQL_End_LE_or_null      = fmt.Sprintf("(%[1]s <= ? OR %[1]s IS NULL)", task_Column_End)
)

func (c *client) ListTasks(search dinkur.SearchTask) ([]dinkur.Task, error) {
	dbTasks, err := c.listDBTasks(search)
	if err != nil {
		return nil, err
	}
	return convTaskSlice(dbTasks), nil
}

func (c *client) listDBTasks(search dinkur.SearchTask) ([]Task, error) {
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
	var dbTasks []Task
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
	if err := q.Find(&dbTasks).Error; err != nil {
		return nil, err
	}
	// we sorted in descending order to get the last tasks.
	// fix this by reversing "again"
	reverseTaskSlice(dbTasks)
	return dbTasks, nil
}

func (c *client) EditTask(edit dinkur.EditTask) (dinkur.UpdatedTask, error) {
	if c.db == nil {
		return dinkur.UpdatedTask{}, ErrNotConnected
	}
	if edit.Name != nil && *edit.Name == "" {
		return dinkur.UpdatedTask{}, ErrTaskNameEmpty
	}
	if edit.Start != nil && edit.End != nil && edit.Start.After(*edit.End) {
		return dinkur.UpdatedTask{}, ErrTaskEndBeforeStart
	}
	var update dinkur.UpdatedTask
	err := c.transaction(func(tx *client) error {
		dbTask, err := tx.getDBTaskToEdit(edit.ID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("no task to edit, failed finding latest task: %w", err)
			}
			return fmt.Errorf("get task to edit: %w", err)
		}
		var anyEdit bool
		update.Old = convTask(dbTask)
		if edit.Name != nil {
			if edit.AppendName {
				dbTask.Name = fmt.Sprint(dbTask.Name, " ", *edit.Name)
			} else {
				dbTask.Name = *edit.Name
			}
			anyEdit = true
		}
		if edit.Start != nil {
			dbTask.Start = *edit.Start
			anyEdit = true
		}
		if edit.End != nil {
			dbTask.End = edit.End
			anyEdit = true
		}
		if dbTask.Elapsed() < 0 {
			return ErrTaskEndBeforeStart
		}
		if anyEdit {
			if err := tx.db.Save(&dbTask).Error; err != nil {
				return fmt.Errorf("save updated task: %w", err)
			}
		}
		update.Updated = convTask(dbTask)
		return nil
	})
	return update, err
}

func (c *client) getDBTaskToEdit(id *uint) (Task, error) {
	var dbTask Task
	err := c.transaction(func(tx *client) error {
		if id != nil {
			dbTaskByID, err := tx.getDBTask(*id)
			if err != nil {
				return fmt.Errorf("get task by ID: %d: %w", *id, err)
			}
			dbTask = dbTaskByID
			return nil
		}
		activeDBTask, err := tx.activeDBTask()
		if err != nil {
			return fmt.Errorf("get active task: %w", err)
		}
		if activeDBTask != nil {
			dbTask = *activeDBTask
			return nil
		}
		now := time.Now()
		dbTasks, err := tx.listDBTasks(dinkur.SearchTask{
			Limit: 1,
			End:   &now,
		})
		if err != nil {
			return fmt.Errorf("list latest 1 task: %w", err)
		}
		if len(dbTasks) == 0 {
			return ErrNotFound
		}
		dbTask = dbTasks[0]
		return nil
	})
	return dbTask, err
}

func (c *client) DeleteTask(id uint) (dinkur.Task, error) {
	if c.db == nil {
		return dinkur.Task{}, ErrNotConnected
	}
	var task dinkur.Task
	err := c.transaction(func(tx *client) error {
		var err error
		task, err = tx.GetTask(id)
		if err != nil {
			return fmt.Errorf("get task to delete: %w", err)
		}
		return tx.db.Delete(&Task{}, id).Error
	})
	return task, err
}

func (c *client) StartTask(task dinkur.NewTask) (dinkur.StartedTask, error) {
	if c.db == nil {
		return dinkur.StartedTask{}, ErrNotConnected
	}
	if task.Name == "" {
		return dinkur.StartedTask{}, ErrTaskNameEmpty
	}
	var start time.Time
	if task.Start != nil {
		start = *task.Start
	} else {
		start = time.Now()
	}
	if task.End != nil && task.End.Before(start) {
		return dinkur.StartedTask{}, ErrTaskEndBeforeStart
	}
	dbTask := Task{
		Name:  task.Name,
		Start: start,
		End:   task.End,
	}
	var startedTask dinkur.StartedTask
	c.transaction(func(tx *client) error {
		var err error
		startedTask.Previous, err = tx.StopActiveTask()
		if err != nil {
			return err
		}
		err = tx.db.Create(&dbTask).Error
		if err != nil {
			return fmt.Errorf("create new active task: %w", err)
		}
		startedTask.New = convTask(dbTask)
		return nil
	})
	return startedTask, nil
}

func (c *client) StopActiveTask() (*dinkur.Task, error) {
	if c.db == nil {
		return nil, ErrNotConnected
	}
	var activeTask *dinkur.Task
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
