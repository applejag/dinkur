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

type NilClient struct {
}

func (c NilClient) Connect() error {
	return ErrClientIsNil
}

func (c NilClient) Ping() error {
	return ErrClientIsNil
}

func (c NilClient) Close() error {
	return ErrClientIsNil
}

func (c NilClient) GetTask(id uint) (Task, error) {
	return Task{}, ErrClientIsNil
}

func (c NilClient) ListTasks(search SearchTask) ([]Task, error) {
	return nil, ErrClientIsNil
}

func (c NilClient) EditTask(edit EditTask) (UpdatedTask, error) {
	return UpdatedTask{}, ErrClientIsNil
}

func (c NilClient) DeleteTask(id uint) (Task, error) {
	return Task{}, ErrClientIsNil
}

func (c NilClient) StartTask(task NewTask) (StartedTask, error) {
	return StartedTask{}, ErrClientIsNil
}

func (c NilClient) ActiveTask() (*Task, error) {
	return nil, ErrClientIsNil
}

func (c NilClient) StopActiveTask() (*Task, error) {
	return nil, ErrClientIsNil
}
