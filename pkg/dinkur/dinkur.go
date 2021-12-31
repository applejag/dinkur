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

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/dinkur/dinkur/pkg/timeutil"
	"gorm.io/gorm"
)

var (
	ErrAlreadyConnected   = errors.New("client is already connected to database")
	ErrNotConnected       = errors.New("client is not connected to database")
	ErrTaskNameEmpty      = errors.New("task name cannot be empty")
	ErrTaskEndBeforeStart = errors.New("task end date cannot be before start date")
	ErrNotFound           = gorm.ErrRecordNotFound
	ErrLimitTooLarge      = errors.New("search limit is too large, maximum: " + strconv.Itoa(math.MaxInt))
	ErrClientIsNil        = errors.New("client is nil")
)

type Client interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	GetTask(ctx context.Context, id uint) (Task, error)
	ListTasks(ctx context.Context, search SearchTask) ([]Task, error)
	EditTask(ctx context.Context, edit EditTask) (UpdatedTask, error)
	DeleteTask(ctx context.Context, id uint) (Task, error)
	StartTask(ctx context.Context, task NewTask) (StartedTask, error)
	ActiveTask(ctx context.Context) (*Task, error)
	StopActiveTask(ctx context.Context) (*Task, error)
}

type SearchTask struct {
	Start *time.Time
	End   *time.Time
	Limit uint

	Shorthand timeutil.TimeSpanShorthand
}

type EditTask struct {
	ID         *uint
	Name       *string
	Start      *time.Time
	End        *time.Time
	AppendName bool
}

type UpdatedTask struct {
	Old     Task
	Updated Task
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

type MigrationStatus byte

const (
	MigrationUnknown MigrationStatus = iota
	MigrationNeverApplied
	MigrationOutdated
	MigrationUpToDate
)

func (s MigrationStatus) String() string {
	switch s {
	case MigrationUnknown:
		return "unknown"
	case MigrationNeverApplied:
		return "never applied"
	case MigrationOutdated:
		return "outdated"
	case MigrationUpToDate:
		return "up to date"
	default:
		return fmt.Sprintf("%T(%d)", s, s)
	}
}
