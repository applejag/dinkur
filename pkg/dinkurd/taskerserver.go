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

func NewTaskerServer(client dinkur.Client) dinkurapiv1.TaskerServer {
	return &taskerServer{client: client}
}

type taskerServer struct {
	dinkurapiv1.UnimplementedTaskerServer
	client dinkur.Client
}

func (c *taskerServer) assertConnected() error {
	if c == nil {
		return ErrTaskerServerIsNil
	}
	if c.client == nil {
		return dinkur.ErrClientIsNil
	}
	return nil
}

func (s *taskerServer) Ping(context.Context, *dinkurapiv1.PingRequest) (*dinkurapiv1.PingResponse, error) {
	if err := s.assertConnected(); err != nil {
		return nil, err
	}
	if err := s.client.Ping(); err != nil {
		return nil, err
	}
	return &dinkurapiv1.PingResponse{}, nil
}

func (s *taskerServer) GetTask(ctx context.Context, req *dinkurapiv1.GetTaskRequest) (*dinkurapiv1.GetTaskResponse, error) {
	if err := s.assertConnected(); err != nil {
		return nil, err
	}
	id, err := uint64ToUint(req.Id)
	if err != nil {
		return nil, err
	}
	task, err := s.client.GetTask(id)
	if err != nil {
		if errors.Is(err, dinkur.ErrNotFound) {
			return &dinkurapiv1.GetTaskResponse{}, nil
		}
		return nil, err
	}
	return &dinkurapiv1.GetTaskResponse{
		Task: convTaskPtr(&task),
	}, nil
}

func (s *taskerServer) GetTaskList(ctx context.Context, req *dinkurapiv1.GetTaskListRequest) (*dinkurapiv1.GetTaskListResponse, error) {
	if err := s.assertConnected(); err != nil {
		return nil, err
	}
	search := dinkur.SearchTask{
		Start:     convTimestampPtr(req.Start),
		End:       convTimestampPtr(req.End),
		Shorthand: convShorthand(req.Shorthand),
	}
	var err error
	search.Limit, err = uint64ToUint(req.Limit)
	if err != nil {
		return nil, err
	}
	tasks, err := s.client.ListTasks(search)
	if err != nil {
		return nil, err
	}
	return &dinkurapiv1.GetTaskListResponse{
		Tasks: convTaskSlice(tasks),
	}, nil
}
