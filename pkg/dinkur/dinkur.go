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

// Package dinkur contains abstractions and models used by multiple Dinkur
// client implementations.
package dinkur

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/dinkur/dinkur/pkg/timeutil"
	"gorm.io/gorm"
)

// Common errors used by multiple Dinkur client and daemon implementations.
var (
	ErrAlreadyConnected   = errors.New("client is already connected to database")
	ErrNotConnected       = errors.New("client is not connected to database")
	ErrTaskNameEmpty      = errors.New("task name cannot be empty")
	ErrTaskEndBeforeStart = errors.New("task end time cannot be before start time")
	ErrNotFound           = gorm.ErrRecordNotFound
	ErrLimitTooLarge      = errors.New("search limit is too large, maximum: " + strconv.Itoa(math.MaxInt))
	ErrClientIsNil        = errors.New("client is nil")
)

// Client is a Dinkur client interface. This is the core interface to act upon
// the Dinkur data store. Depending on the implementation, it may either talk
// directly to an Sqlite3 database file, or talk to a Dinkur daemon via gRPC
// over TCP/IP that in turn talks to a database.
type Client interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	Tasker
	Alerter
}

// Tasker is the Dinkur client methods targeted to reading, creating, and
// updating tasks.
type Tasker interface {
	GetTask(ctx context.Context, id uint) (Task, error)
	ListTasks(ctx context.Context, search SearchTask) ([]Task, error)
	EditTask(ctx context.Context, edit EditTask) (UpdatedTask, error)
	DeleteTask(ctx context.Context, id uint) (Task, error)
	StartTask(ctx context.Context, task NewTask) (StartedTask, error)
	ActiveTask(ctx context.Context) (*Task, error)
	StopActiveTask(ctx context.Context, endTime time.Time) (*Task, error)
}

// Alerter is the Dinkur client methods targeted to reading alerts.
type Alerter interface {
	StreamAlert(ctx context.Context) (<-chan StreamedAlert, error)
	GetAlertList(ctx context.Context) ([]Alert, error)
	DeleteAlert(ctx context.Context, id uint) (Alert, error)
}

// StreamedAlert holds an alert and its event type.
type StreamedAlert struct {
	Alert Alert
	Event EventType
}

// SearchTask holds parameters used when searching for list of tasks.
type SearchTask struct {
	Start *time.Time
	End   *time.Time
	Limit uint

	Shorthand          timeutil.TimeSpanShorthand
	NameFuzzy          string
	NameHighlightStart string
	NameHighlightEnd   string
}

// EditTask holds parameters used when editing a task.
type EditTask struct {
	// IDOrZero of the task to edit. If set to nil, then Dinkur will attempt to make
	// an educated guess on what task to edit by editing the active task or a
	// recent task.
	IDOrZero uint
	// Name is the new task name. If AppendName is enabled, then this value will
	// append to the existing name, delimited with a space.
	//
	// No change to the task name is applied if this is set to nil.
	Name *string
	// Start is the new task start timestamp.
	//
	// No change to the task start timestamp is applied if this is set to nil.
	Start *time.Time
	// End is the new task end timestamp.
	//
	// No change to the task end timestamp is applied if this is set to nil.
	End *time.Time
	// AppendName changes the name field to append the name to the task's
	// existing name (delimited with a space) instead of replacing it.
	AppendName         bool
	StartAfterIDOrZero uint
	EndBeforeIDOrZero  uint
	StartAfterLast     bool
}

// UpdatedTask is the response from an edited task, with values for before the
// edits were applied and after they were applied.
type UpdatedTask struct {
	Old     Task
	Updated Task
}

// NewTask holds parameters used when creating a new task.
type NewTask struct {
	Name               string
	Start              *time.Time
	End                *time.Time
	StartAfterIDOrZero uint
	EndBeforeIDOrZero  uint
	StartAfterLast     bool
}

// StartedTask is the response from creating a new task, with the newly created
// task object as well as the task that was stopped when creating the task,
// if any task was previously active.
type StartedTask struct {
	New      Task
	Previous *Task
}
