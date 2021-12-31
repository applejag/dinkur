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
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// CommonFields contains fields used by multiple other models.
type CommonFields struct {
	// ID is a unique identifier for this task. The same ID will never be used
	// twice for a given database.
	ID uint `gorm:"primarykey"`
	// CreatedAt stores when the database object was added to the database.
	//
	// It is automatically set by GORM due to its naming convention.
	CreatedAt time.Time
	// UpdatedAt stores when the database object was added to the database or
	// when it was last updated.
	//
	// It is automatically set by GORM due to its naming convention.
	UpdatedAt time.Time
}

func convCommonFields(f CommonFields) dinkur.CommonFields {
	return dinkur.CommonFields{
		ID:        f.ID,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

const (
	taskFieldEnd = "End"

	taskColumnStart = "start"
	taskColumnEnd   = "end"
)

// Task is a time tracked task stored in the database.
type Task struct {
	CommonFields
	// Name of the task.
	Name string `gorm:"not null;default:''"`
	// Start time of the task.
	Start time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	// End time of the task, or nil if the task is still active.
	End *time.Time `gorm:"index"`
}

// Elapsed returns the duration of the task. If the task is currently active,
// the duration is calculated from the start to now.
func (t Task) Elapsed() time.Duration {
	var end time.Time
	if t.End != nil {
		end = *t.End
	} else {
		end = time.Now()
	}
	return end.Sub(t.Start)
}

func timePtrUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utcTime := (*t).UTC()
	return &utcTime
}

func timePtrLocal(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utcTime := (*t).Local()
	return &utcTime
}

func convTask(t Task) dinkur.Task {
	return dinkur.Task{
		CommonFields: convCommonFields(t.CommonFields),
		Name:         t.Name,
		Start:        t.Start.Local(),
		End:          timePtrLocal(t.End),
	}
}

func convTaskPtr(t *Task) *dinkur.Task {
	if t == nil {
		return nil
	}
	dinkurTask := convTask(*t)
	return &dinkurTask
}

func convTaskSlice(slice []Task) []dinkur.Task {
	result := make([]dinkur.Task, len(slice))
	for i, t := range slice {
		result[i] = convTask(t)
	}
	return result
}

// Migration holds the latest migration revision identifier. At most one row of
// this object is expected to be in the database at any given time.
type Migration struct {
	CommonFields
	Version int
}
