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
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

func (c *client) ActiveTask() (*dinkur.Task, error) {
	dbTask, err := c.activeDBTask()
	if err != nil {
		return nil, err
	}
	return convTaskPtr(dbTask), nil
}

func (c *client) activeDBTask() (*Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
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
	if err := c.assertConnected(); err != nil {
		return Task{}, err
	}
	var dbTask Task
	err := c.db.First(&dbTask, id).Error
	if err != nil {
		return Task{}, err
	}
	return dbTask, nil
}

var (
	task_SQL_Between_Start = fmt.Sprintf(
		"(%[1]s >= @start) OR "+
			"(%[2]s IS NOT NULL AND %[1]s >= @start) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP >= @start)",
		task_Column_Start, task_Column_End,
	)

	task_SQL_Between_End = fmt.Sprintf(
		"(%[2]s <= @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s <= @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP <= @end)",
		task_Column_Start, task_Column_End,
	)

	task_SQL_Between = fmt.Sprintf(
		"(%[1]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP BETWEEN @start AND @end)",
		task_Column_Start, task_Column_End,
	)
)

func (c *client) ListTasks(search dinkur.SearchTask) ([]dinkur.Task, error) {
	dbTasks, err := c.listDBTasks(search)
	if err != nil {
		return nil, err
	}
	return convTaskSlice(dbTasks), nil
}

func (c *client) listDBTasks(search dinkur.SearchTask) ([]Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	span := search.Shorthand.Span(time.Now())
	if search.Start == nil {
		search.Start = span.Start
	}
	if search.End == nil {
		search.End = span.End
	}
	if search.Limit > math.MaxInt {
		return nil, dinkur.ErrLimitTooLarge
	}
	var dbTasks []Task
	q := c.db.Model(&Task{}).
		Order(task_Column_Start + " desc").
		Limit(int(search.Limit))
	switch {
	case search.Start != nil && search.End != nil:
		// adding/subtracting 1s to resolve rounding issues, as Sqlite's
		// smallest time unit is a second.
		start := (*search.Start).UTC().Add(-time.Second)
		end := (*search.End).UTC().Add(time.Second)
		q = q.Or(c.db.Where(task_SQL_Between, sql.Named("start", start), sql.Named("end", end)))
	case search.Start != nil:
		start := (*search.Start).UTC().Add(-time.Second)
		q = q.Or(c.db.Where(task_SQL_Between_Start, sql.Named("start", start)))
	case search.End != nil:
		end := (*search.End).UTC().Add(time.Second)
		q = q.Or(c.db.Where(task_SQL_Between_End, sql.Named("end", end)))
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
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedTask{}, err
	}
	if edit.Name != nil && *edit.Name == "" {
		return dinkur.UpdatedTask{}, dinkur.ErrTaskNameEmpty
	}
	if edit.Start != nil && edit.End != nil && edit.Start.After(*edit.End) {
		return dinkur.UpdatedTask{}, dinkur.ErrTaskEndBeforeStart
	}
	var update dinkur.UpdatedTask
	err := c.transaction(func(tx *client) error {
		dbTask, err := tx.getDBTaskToEdit(edit.ID)
		if err != nil {
			if errors.Is(err, dinkur.ErrNotFound) {
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
			return dinkur.ErrTaskEndBeforeStart
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
			return dinkur.ErrNotFound
		}
		dbTask = dbTasks[0]
		return nil
	})
	return dbTask, err
}

func (c *client) DeleteTask(id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
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
	if err := c.assertConnected(); err != nil {
		return dinkur.StartedTask{}, err
	}
	if task.Name == "" {
		return dinkur.StartedTask{}, dinkur.ErrTaskNameEmpty
	}
	var start time.Time
	if task.Start != nil {
		start = *task.Start
	} else {
		start = time.Now()
	}
	if task.End != nil && task.End.Before(start) {
		return dinkur.StartedTask{}, dinkur.ErrTaskEndBeforeStart
	}
	dbTask := Task{
		Name:  task.Name,
		Start: start.UTC(),
		End:   timePtrUTC(task.End),
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
	if err := c.assertConnected(); err != nil {
		return nil, err
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
