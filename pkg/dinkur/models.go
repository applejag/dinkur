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

import "time"

// CommonFields contains fields used by multiple other models.
type CommonFields struct {
	// ID is a unique identifier for this task. The same ID will never be used
	// twice for a given database.
	ID uint `json:"id"`
	// CreatedAt is when the object was created.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores when the object was last updated/edited.
	UpdatedAt time.Time `json:"updatedAt"`
}

// Task is a time tracked task.
type Task struct {
	CommonFields
	// Name of the task.
	Name string `json:"name"`
	// Start time of the task.
	Start time.Time `json:"start"`
	// End time of the task, or nil if the task is still active.
	End *time.Time `json:"end"`
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
