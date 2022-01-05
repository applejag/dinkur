// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
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

package dinkurclient

import (
	"context"
	"fmt"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

func (c *client) GetTask(ctx context.Context, id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
	}
	res, err := c.tasker.GetTask(ctx, &dinkurapiv1.GetTaskRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Task{}, convError(err)
	}
	if res == nil {
		return dinkur.Task{}, ErrResponseIsNil
	}
	task, err := convTaskPtrNoNil(res.Task)
	if err != nil {
		return dinkur.Task{}, convError(err)
	}
	return task, nil
}

func (c *client) ListTasks(ctx context.Context, search dinkur.SearchTask) ([]dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	req := dinkurapiv1.GetTaskListRequest{
		Start:     convTimePtr(search.Start),
		End:       convTimePtr(search.End),
		Limit:     uint64(search.Limit),
		Shorthand: convShorthand(search.Shorthand),
		NameFuzzy: search.NameFuzzy,
	}
	res, err := c.tasker.GetTaskList(ctx, &req)
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	tasks, err := convTaskSlice(res.Tasks)
	if err != nil {
		return nil, convError(err)
	}
	return tasks, nil
}

func (c *client) EditTask(ctx context.Context, edit dinkur.EditTask) (dinkur.UpdatedTask, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedTask{}, err
	}
	res, err := c.tasker.UpdateTask(ctx, &dinkurapiv1.UpdateTaskRequest{
		IdOrZero:           uint64(edit.IDOrZero),
		Name:               convStringPtr(edit.Name),
		Start:              convTimePtr(edit.Start),
		End:                convTimePtr(edit.End),
		AppendName:         edit.AppendName,
		StartAfterIdOrZero: uint64(edit.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(edit.EndBeforeIDOrZero),
		StartAfterLast:     edit.StartAfterLast,
	})
	if err != nil {
		return dinkur.UpdatedTask{}, convError(err)
	}
	if res == nil {
		return dinkur.UpdatedTask{}, ErrResponseIsNil
	}
	taskBefore, err := convTaskPtrNoNil(res.Before)
	if err != nil {
		return dinkur.UpdatedTask{}, fmt.Errorf("task before: %w", convError(err))
	}
	taskAfter, err := convTaskPtrNoNil(res.After)
	if err != nil {
		return dinkur.UpdatedTask{}, fmt.Errorf("task after: %w", convError(err))
	}
	return dinkur.UpdatedTask{
		Old:     taskBefore,
		Updated: taskAfter,
	}, nil
}

func (c *client) DeleteTask(ctx context.Context, id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
	}
	res, err := c.tasker.DeleteTask(ctx, &dinkurapiv1.DeleteTaskRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Task{}, convError(err)
	}
	if res == nil {
		return dinkur.Task{}, ErrResponseIsNil
	}
	task, err := convTaskPtrNoNil(res.DeletedTask)
	if err != nil {
		return dinkur.Task{}, convError(err)
	}
	return task, nil
}

func (c *client) StartTask(ctx context.Context, task dinkur.NewTask) (dinkur.StartedTask, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.StartedTask{}, err
	}
	res, err := c.tasker.CreateTask(ctx, &dinkurapiv1.CreateTaskRequest{
		Name:               task.Name,
		Start:              convTimePtr(task.Start),
		End:                convTimePtr(task.End),
		StartAfterIdOrZero: uint64(task.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(task.EndBeforeIDOrZero),
		StartAfterLast:     task.StartAfterLast,
	})
	if err != nil {
		return dinkur.StartedTask{}, convError(err)
	}
	if res == nil {
		return dinkur.StartedTask{}, ErrResponseIsNil
	}
	prevTask, err := convTaskPtr(res.PreviouslyActiveTask)
	if err != nil {
		return dinkur.StartedTask{}, fmt.Errorf("stopped task: %w", convError(err))
	}
	newTask, err := convTaskPtrNoNil(res.CreatedTask)
	if err != nil {
		return dinkur.StartedTask{}, fmt.Errorf("created task: %w", convError(err))
	}
	return dinkur.StartedTask{
		Previous: prevTask,
		New:      newTask,
	}, nil
}

func (c *client) ActiveTask(ctx context.Context) (*dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.tasker.GetActiveTask(ctx, &dinkurapiv1.GetActiveTaskRequest{})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	task, err := convTaskPtr(res.ActiveTask)
	if err != nil {
		return nil, convError(err)
	}
	return task, nil
}

func (c *client) StopActiveTask(ctx context.Context, endTime time.Time) (*dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.tasker.StopActiveTask(ctx, &dinkurapiv1.StopActiveTaskRequest{
		End: convTimePtr(&endTime),
	})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	task, err := convTaskPtr(res.StoppedTask)
	if err != nil {
		return nil, convError(err)
	}
	return task, nil
}
