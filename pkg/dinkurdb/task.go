// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package dinkurdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
)

func (c *client) ActiveTask(ctx context.Context) (*dinkur.Task, error) {
	dbTask, err := c.withContext(ctx).activeDBTask()
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
	err := c.db.Where(Task{End: nil}, taskFieldEnd).First(&dbTask).Error
	if err != nil {
		return nil, nilNotFoundError(err)
	}
	return &dbTask, nil
}

func (c *client) GetTask(ctx context.Context, id uint) (dinkur.Task, error) {
	dbTask, err := c.withContext(ctx).getDBTask(id)
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
	taskSQLBetweenStart = fmt.Sprintf(
		"((%[1]s >= @start) OR "+
			"(%[2]s IS NOT NULL AND %[1]s >= @start) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP >= @start))",
		taskColumnStart, taskColumnEnd,
	)

	taskSQLBetweenEnd = fmt.Sprintf(
		"((%[2]s <= @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s <= @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP <= @end))",
		taskColumnStart, taskColumnEnd,
	)

	taskSQLBetween = fmt.Sprintf(
		"((%[1]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP BETWEEN @start AND @end))",
		taskColumnStart, taskColumnEnd,
	)
)

func (c *client) ListTasks(ctx context.Context, search dinkur.SearchTask) ([]dinkur.Task, error) {
	dbTasks, err := c.withContext(ctx).listDBTasks(search)
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
		Order(taskColumnStart + " DESC").
		Limit(int(search.Limit))
	switch {
	case search.Start != nil && search.End != nil:
		// adding/subtracting 1s to resolve rounding issues, as Sqlite's
		// smallest time unit is a second.
		start := (*search.Start).UTC().Add(-time.Second)
		end := (*search.End).UTC().Add(time.Second)
		q = q.Where(taskSQLBetween, sql.Named("start", start), sql.Named("end", end))
	case search.Start != nil:
		start := (*search.Start).UTC().Add(-time.Second)
		q = q.Where(taskSQLBetweenStart, sql.Named("start", start))
	case search.End != nil:
		end := (*search.End).UTC().Add(time.Second)
		q = q.Where(taskSQLBetweenEnd, sql.Named("end", end))
	}
	if search.NameFuzzy != "" {
		if search.NameHighlightStart != "" || search.NameHighlightEnd != "" {
			q = q.Joins("INNER JOIN tasks_idx ON tasks.id = tasks_idx.rowid").
				Select(
					"id, created_at, updated_at, highlight(tasks_idx, 0, ?, ?) AS name, start, end",
					search.NameHighlightStart, search.NameHighlightEnd).
				Where(taskFTS5ColumnName+" MATCH ?", search.NameFuzzy)
		} else {
			subQ := c.db.Model(&TaskFTS5{}).
				Select(taskFTS5ColumnRowID).
				Where(taskFTS5ColumnName+" MATCH ?", search.NameFuzzy)
			q = q.Where(taskColumnID+" IN (?)", subQ)
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

func (c *client) EditTask(ctx context.Context, edit dinkur.EditTask) (dinkur.UpdatedTask, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedTask{}, err
	}
	update, err := c.withContext(ctx).editDBTask(edit)
	if err != nil {
		return dinkur.UpdatedTask{}, err
	}
	c.taskObs.pubTaskWait(taskEvent{
		dbTask: update.updated,
		event:  dinkur.EventUpdated,
	})
	return dinkur.UpdatedTask{
		Old:     convTask(update.old),
		Updated: convTask(update.updated),
	}, nil
}

type updatedDBTask struct {
	old     Task
	updated Task
}

func (c *client) editDBTask(edit dinkur.EditTask) (updatedDBTask, error) {
	if edit.Name != nil && *edit.Name == "" {
		return updatedDBTask{}, dinkur.ErrTaskNameEmpty
	}
	if edit.Start != nil && edit.End != nil && edit.Start.After(*edit.End) {
		return updatedDBTask{}, dinkur.ErrTaskEndBeforeStart
	}
	var update updatedDBTask
	err := c.transaction(func(tx *client) (tranErr error) {
		update, tranErr = tx.editDBTaskNoTran(edit)
		return
	})
	return update, err
}

func (c *client) editDBTaskNoTran(edit dinkur.EditTask) (updatedDBTask, error) {
	dbTask, err := c.getDBTaskToEditNoTran(edit.IDOrZero)
	if err != nil {
		if errors.Is(err, dinkur.ErrNotFound) {
			return updatedDBTask{}, fmt.Errorf("no task to edit, failed finding latest task: %w", err)
		}
		return updatedDBTask{}, fmt.Errorf("get task to edit: %w", err)
	}
	startAfterTime, err := c.getTimeToStartAfterOrNow(edit.StartAfterIDOrZero, edit.StartAfterLast)
	if err != nil {
		return updatedDBTask{}, err
	}
	if startAfterTime != nil {
		edit.Start = startAfterTime
	}
	endBeforeTime, err := c.getTimeToEndBefore(edit.EndBeforeIDOrZero)
	if err != nil {
		return updatedDBTask{}, err
	}
	if endBeforeTime != nil {
		edit.End = endBeforeTime
	}
	var anyEdit bool
	taskBeforeEdit := dbTask
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
		return updatedDBTask{}, dinkur.ErrTaskEndBeforeStart
	}
	if anyEdit {
		if err := c.db.Save(&dbTask).Error; err != nil {
			return updatedDBTask{}, fmt.Errorf("save updated task: %w", err)
		}
	}
	return updatedDBTask{
		old:     taskBeforeEdit,
		updated: dbTask,
	}, nil
}

func (c *client) getDBTaskToStartAfter(idOrZero uint, lastTask bool) (*Task, error) {
	if idOrZero != 0 {
		startAfter, err := c.getDBTask(idOrZero)
		if err != nil {
			return nil, fmt.Errorf("get task by ID to start after: %w", err)
		}
		return &startAfter, nil
	} else if lastTask {
		lastTasks, err := c.listDBTasks(dinkur.SearchTask{
			Shorthand: timeutil.TimeSpanNone,
			Limit:     1,
		})
		if err != nil {
			return nil, fmt.Errorf("get last task to start after: %w", err)
		}
		if len(lastTasks) == 0 {
			return nil, fmt.Errorf("get last task to start after: %w", dinkur.ErrNotFound)
		}
		return &lastTasks[0], nil
	}
	return nil, nil
}

func (c *client) getTimeToStartAfterOrNow(idOrZero uint, lastTask bool) (*time.Time, error) {
	startAfter, err := c.getDBTaskToStartAfter(idOrZero, lastTask)
	if err != nil {
		return nil, err
	}
	if startAfter == nil {
		return nil, nil
	}
	if startAfter.End == nil {
		now := time.Now()
		return &now, nil
	}
	return startAfter.End, nil
}

func (c *client) getDBTaskToEndBefore(idOrZero uint) (*Task, error) {
	if idOrZero == 0 {
		return nil, nil
	}
	endBefore, err := c.getDBTask(idOrZero)
	if err != nil {
		return nil, fmt.Errorf("get task by ID to end before: %w", err)
	}
	return &endBefore, nil
}

func (c *client) getTimeToEndBefore(idOrZero uint) (*time.Time, error) {
	endBefore, err := c.getDBTaskToEndBefore(idOrZero)
	if err != nil {
		return nil, err
	}
	if endBefore == nil {
		return nil, nil
	}
	return &endBefore.Start, nil
}

func (c *client) getDBTaskToEditNoTran(idOrZero uint) (Task, error) {
	if idOrZero != 0 {
		dbTaskByID, err := c.getDBTask(idOrZero)
		if err != nil {
			return Task{}, fmt.Errorf("get task by ID: %d: %w", idOrZero, err)
		}
		return dbTaskByID, nil
	}
	activeDBTask, err := c.activeDBTask()
	if err != nil {
		return Task{}, fmt.Errorf("get active task: %w", err)
	}
	if activeDBTask != nil {
		return *activeDBTask, nil
	}
	now := time.Now()
	dbTasks, err := c.listDBTasks(dinkur.SearchTask{
		Limit: 1,
		End:   &now,
	})
	if err != nil {
		return Task{}, fmt.Errorf("list latest 1 task: %w", err)
	}
	if len(dbTasks) == 0 {
		return Task{}, dinkur.ErrNotFound
	}
	return dbTasks[0], nil
}

func (c *client) DeleteTask(ctx context.Context, id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
	}
	dbTask, err := c.withContext(ctx).deleteDBTask(id)
	if err != nil {
		return dinkur.Task{}, err
	}
	c.taskObs.pubTaskWait(taskEvent{
		dbTask: dbTask,
		event:  dinkur.EventDeleted,
	})
	return convTask(dbTask), err
}

func (c *client) deleteDBTask(id uint) (Task, error) {
	var dbTask Task
	err := c.transaction(func(tx *client) (tranErr error) {
		dbTask, tranErr = tx.deleteDBTaskNoTran(id)
		return
	})
	return dbTask, err
}

func (c *client) deleteDBTaskNoTran(id uint) (Task, error) {
	dbTask, err := c.getDBTask(id)
	if err != nil {
		return Task{}, fmt.Errorf("get task to delete: %w", err)
	}
	if err := c.db.Delete(&Task{}, id).Error; err != nil {
		return Task{}, fmt.Errorf("delete task: %w", err)
	}
	return dbTask, nil
}

func (c *client) StartTask(ctx context.Context, task dinkur.NewTask) (dinkur.StartedTask, error) {
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
	newTask := newTask{
		Task: Task{
			Name:  task.Name,
			Start: start.UTC(),
			End:   timePtrUTC(task.End),
		},
		startAfterIDOrZero: task.StartAfterIDOrZero,
		endBeforeIDOrZero:  task.EndBeforeIDOrZero,
		startAfterLast:     task.StartAfterLast,
	}
	startedTask, err := c.withContext(ctx).startDBTask(newTask)
	if err != nil {
		return dinkur.StartedTask{}, err
	}
	if startedTask.previous != nil {
		c.taskObs.pubTaskWait(taskEvent{
			dbTask: *startedTask.previous,
			event:  dinkur.EventUpdated,
		})
	}
	c.taskObs.pubTaskWait(taskEvent{
		dbTask: startedTask.new,
		event:  dinkur.EventCreated,
	})
	return dinkur.StartedTask{
		New:      convTask(startedTask.new),
		Previous: convTaskPtr(startedTask.previous),
	}, nil
}

type startedDBTask struct {
	previous *Task
	new      Task
}

type newTask struct {
	Task
	startAfterIDOrZero uint
	endBeforeIDOrZero  uint
	startAfterLast     bool
}

func (c *client) startDBTask(newTask newTask) (startedDBTask, error) {
	var startedTask startedDBTask
	err := c.transaction(func(tx *client) (tranErr error) {
		startedTask, tranErr = tx.startDBTaskNoTran(newTask)
		return
	})
	return startedTask, err
}

func (c *client) startDBTaskNoTran(newTask newTask) (startedDBTask, error) {
	startAfterTime, err := c.getTimeToStartAfterOrNow(newTask.startAfterIDOrZero, newTask.startAfterLast)
	if err != nil {
		return startedDBTask{}, err
	}
	if startAfterTime != nil {
		newTask.Start = *startAfterTime
	}
	endBeforeTime, err := c.getTimeToEndBefore(newTask.endBeforeIDOrZero)
	if err != nil {
		return startedDBTask{}, err
	}
	if endBeforeTime != nil {
		newTask.End = endBeforeTime
	}
	previousDBTask, err := c.stopActiveDBTaskNoTran(newTask.Start)
	if err != nil {
		return startedDBTask{}, fmt.Errorf("stop previously active task: %w", err)
	}
	err = c.db.Create(&newTask.Task).Error
	if err != nil {
		return startedDBTask{}, fmt.Errorf("create new active task: %w", err)
	}
	return startedDBTask{
		previous: previousDBTask,
		new:      newTask.Task,
	}, nil
}

func (c *client) StopActiveTask(ctx context.Context, endTime time.Time) (*dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbTask, err := c.withContext(ctx).stopActiveDBTask(endTime)
	if err != nil {
		return nil, err
	}
	if err == nil && dbTask != nil {
		c.taskObs.pubTaskWait(taskEvent{
			dbTask: *dbTask,
			event:  dinkur.EventUpdated,
		})
	}
	return convTaskPtr(dbTask), nil
}

func (c *client) stopActiveDBTask(endTime time.Time) (*Task, error) {
	var activeDBTask *Task
	err := c.transaction(func(tx *client) (tranErr error) {
		activeDBTask, tranErr = tx.stopActiveDBTaskNoTran(endTime)
		return
	})
	return activeDBTask, err
}

func (c *client) stopActiveDBTaskNoTran(endTime time.Time) (*Task, error) {
	var tasks []Task
	if err := c.db.Where(&Task{End: nil}, taskFieldEnd).Find(&tasks).Error; err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	for i, task := range tasks {
		if endTime.Before(task.Start) {
			return nil, dinkur.ErrTaskEndBeforeStart
		}
		tasks[i].End = &endTime
	}
	err := c.db.Model(&Task{}).
		Where(&Task{End: nil}, taskFieldEnd).
		Update(taskFieldEnd, endTime).
		Error
	if err != nil {
		return nil, err
	}
	return &tasks[0], nil
}

func (c *client) StreamTask(ctx context.Context) (<-chan dinkur.StreamedTask, error) {
	ch := make(chan dinkur.StreamedTask, 4)
	go c.streamTaskGoroutine(ctx, ch)
	return ch, nil
}

func (c *client) streamTaskGoroutine(ctx context.Context, ch chan dinkur.StreamedTask) {
	done := ctx.Done()
	dbTaskChan := c.taskObs.subTasks()
	defer close(ch)
	defer c.taskObs.unsubTasks(dbTaskChan)
	for {
		select {
		case ev, ok := <-dbTaskChan:
			if !ok {
				return
			}
			ch <- dinkur.StreamedTask{
				Task:  convTask(ev.dbTask),
				Event: ev.event,
			}
		case <-done:
			return
		}
	}
}

func reverseTaskSlice(slice []Task) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
