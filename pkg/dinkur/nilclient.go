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

package dinkur

import "context"

type NilClient struct {
}

func (NilClient) Connect(context.Context) error {
	return ErrClientIsNil
}

func (NilClient) Ping(context.Context) error {
	return ErrClientIsNil
}

func (NilClient) Close() error {
	return ErrClientIsNil
}

func (NilClient) GetTask(context.Context, uint) (Task, error) {
	return Task{}, ErrClientIsNil
}

func (NilClient) ListTasks(context.Context, SearchTask) ([]Task, error) {
	return nil, ErrClientIsNil
}

func (NilClient) EditTask(context.Context, EditTask) (UpdatedTask, error) {
	return UpdatedTask{}, ErrClientIsNil
}

func (NilClient) DeleteTask(context.Context, uint) (Task, error) {
	return Task{}, ErrClientIsNil
}

func (NilClient) StartTask(context.Context, NewTask) (StartedTask, error) {
	return StartedTask{}, ErrClientIsNil
}

func (NilClient) ActiveTask(context.Context) (*Task, error) {
	return nil, ErrClientIsNil
}

func (NilClient) StopActiveTask(context.Context) (*Task, error) {
	return nil, ErrClientIsNil
}
