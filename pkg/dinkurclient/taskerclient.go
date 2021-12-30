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

	"github.com/dinkur/dinkur/pkg/dinkur"
	"google.golang.org/grpc"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

type Options struct{}

func NewClient(serverAddr string, opt Options) dinkur.Client {
	return &client{
		Options:    opt,
		serverAddr: serverAddr,
	}
}

type client struct {
	Options
	serverAddr string
	conn       *grpc.ClientConn
	tasker     dinkurapiv1.TaskerClient
}

func (c *client) assertConnected() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn == nil || c.tasker == nil {
		return dinkur.ErrNotConnected
	}
	return nil
}

func (c *client) Connect() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn != nil || c.tasker != nil {
		return dinkur.ErrAlreadyConnected
	}
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure()) // TODO: add credentials via opts args
	if err != nil {
		return err
	}
	c.conn = conn
	c.tasker = dinkurapiv1.NewTaskerClient(conn)
	return nil
}

func (c *client) Close() (err error) {
	if conn := c.conn; conn != nil {
		err = conn.Close()
		c.conn = nil
	}
	c.tasker = nil
	return
}

func (c *client) Ping() error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	res, err := c.tasker.Ping(context.TODO(), &dinkurapiv1.PingRequest{})
	if err != nil {
		return err
	}
	if res == nil {
		return ErrResponseIsNil
	}
	return nil
}

func (c *client) GetTask(id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
	}
	return dinkur.Task{}, ErrNotImplemented
}

func (c *client) ListTasks(search dinkur.SearchTask) ([]dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	req := dinkurapiv1.GetTaskListRequest{
		Start:     convTimePtr(search.Start),
		End:       convTimePtr(search.End),
		Limit:     uint64(search.Limit),
		Shorthand: convShorthand(search.Shorthand),
	}
	res, err := c.tasker.GetTaskList(context.TODO(), &req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	tasks, err := convTaskSlice(res.Tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *client) EditTask(edit dinkur.EditTask) (dinkur.UpdatedTask, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedTask{}, err
	}
	return dinkur.UpdatedTask{}, ErrNotImplemented
}

func (c *client) DeleteTask(id uint) (dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Task{}, err
	}
	return dinkur.Task{}, ErrNotImplemented
}

func (c *client) StartTask(task dinkur.NewTask) (dinkur.StartedTask, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.StartedTask{}, err
	}
	return dinkur.StartedTask{}, ErrNotImplemented
}

func (c *client) ActiveTask() (*dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	return nil, ErrNotImplemented
}

func (c *client) StopActiveTask() (*dinkur.Task, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	return nil, ErrNotImplemented
}
