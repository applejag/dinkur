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

package dinkurd

import (
	"context"
	"errors"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

func (d *daemon) Ping(ctx context.Context, req *dinkurapiv1.PingRequest) (*dinkurapiv1.PingResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	if err := d.client.Ping(ctx); err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.PingResponse{}, nil
}

func (d *daemon) GetTask(ctx context.Context, req *dinkurapiv1.GetTaskRequest) (*dinkurapiv1.GetTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	id, err := uint64ToUint(req.Id)
	if err != nil {
		return nil, convError(err)
	}
	task, err := d.client.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, dinkur.ErrNotFound) {
			return &dinkurapiv1.GetTaskResponse{}, nil
		}
		return nil, convError(err)
	}
	return &dinkurapiv1.GetTaskResponse{
		Task: convTaskPtr(&task),
	}, nil
}

func (d *daemon) GetActiveTask(ctx context.Context, req *dinkurapiv1.GetActiveTaskRequest) (*dinkurapiv1.GetActiveTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	task, err := d.client.ActiveTask(ctx)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.GetActiveTaskResponse{
		ActiveTask: convTaskPtr(task),
	}, nil
}

func (d *daemon) GetTaskList(ctx context.Context, req *dinkurapiv1.GetTaskListRequest) (*dinkurapiv1.GetTaskListResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	search := dinkur.SearchTask{
		Start:     convTimestampPtr(req.Start),
		End:       convTimestampPtr(req.End),
		Shorthand: convShorthand(req.Shorthand),
	}
	var err error
	search.Limit, err = uint64ToUint(req.Limit)
	if err != nil {
		return nil, convError(err)
	}
	tasks, err := d.client.ListTasks(ctx, search)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.GetTaskListResponse{
		Tasks: convTaskSlice(tasks),
	}, nil
}

func (d *daemon) CreateTask(ctx context.Context, req *dinkurapiv1.CreateTaskRequest) (*dinkurapiv1.CreateTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	newTask := dinkur.NewTask{
		Name:  req.Name,
		Start: convTimestampPtr(req.Start),
		End:   convTimestampPtr(req.End),
	}
	startedTask, err := d.client.StartTask(ctx, newTask)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.CreateTaskResponse{
		PreviouslyActiveTask: convTaskPtr(startedTask.Previous),
		CreatedTask:          convTaskPtr(&startedTask.New),
	}, nil
}

func (d *daemon) UpdateTask(ctx context.Context, req *dinkurapiv1.UpdateTaskRequest) (*dinkurapiv1.UpdateTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	id, err := convUint64(req.Id)
	if err != nil {
		return nil, convError(err)
	}
	edit := dinkur.EditTask{
		Name:       convString(req.Name),
		Start:      convTimestampPtr(req.Start),
		End:        convTimestampPtr(req.End),
		ID:         id,
		AppendName: req.AppendName,
	}
	update, err := d.client.EditTask(ctx, edit)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.UpdateTaskResponse{
		Before: convTaskPtr(&update.Old),
		After:  convTaskPtr(&update.Updated),
	}, nil
}

func (d *daemon) DeleteTask(ctx context.Context, req *dinkurapiv1.DeleteTaskRequest) (*dinkurapiv1.DeleteTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	id, err := uint64ToUint(req.Id)
	if err != nil {
		return nil, convError(err)
	}
	deletedTask, err := d.client.DeleteTask(ctx, id)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.DeleteTaskResponse{
		DeletedTask: convTaskPtr(&deletedTask),
	}, nil
}

func (d *daemon) StopActiveTask(ctx context.Context, req *dinkurapiv1.StopActiveTaskRequest) (*dinkurapiv1.StopActiveTaskResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	stoppedTask, err := d.client.StopActiveTask(ctx, convTimestampOrNow(req.End))
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.StopActiveTaskResponse{
		StoppedTask: convTaskPtr(stoppedTask),
	}, nil
}
