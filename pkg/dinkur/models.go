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

// EventType is the type of a streamed event.
type EventType byte

const (
	// EventCreated means the subject was just created.
	EventCreated EventType = iota
	// EventUpdated means the subject was just updated.
	EventUpdated
	// EventDeleted means the subject was just deleted.
	EventDeleted
)

// Alert is a notfication provided by Dinkur, such as an alert when the user
// has gone AFK.
type Alert struct {
	CommonFields
	Type alertUnion
}

type alertUnion interface {
	isAlertUnion()
}

// PlainMessage returns the inner plain message alert, or false if the alert
// is of a different type.
func (a Alert) PlainMessage() (AlertPlainMessage, bool) {
	if inner, ok := a.Type.(AlertPlainMessage); ok {
		return inner, true
	}
	return AlertPlainMessage{}, false
}

// AFK returns the inner AFK alert, or false if the alert is of a different type.
func (a Alert) AFK() (AlertAFK, bool) {
	if inner, ok := a.Type.(AlertAFK); ok {
		return inner, true
	}
	return AlertAFK{}, false
}

// FormerlyAFK returns the inner formerly AFK alert, or false if the alert
// is of a different type.
func (a Alert) FormerlyAFK() (AlertFormerlyAFK, bool) {
	if inner, ok := a.Type.(AlertFormerlyAFK); ok {
		return inner, true
	}
	return AlertFormerlyAFK{}, false
}

// AlertPlainMessage is a type of alert for generic messages that needs to be
// presented to the user with no need for user action.
type AlertPlainMessage struct {
	alertUnion
	Message string
}

func (AlertPlainMessage) isAlertUnion() {}

// AlertAFK is a type of alert that's issued when the user has just become AFK
// (away from keyboard).
type AlertAFK struct {
	alertUnion
	ActiveTask Task
}

func (AlertAFK) isAlertUnion() {}

// AlertFormerlyAFK is a type of alert that's issued when the user is no longer
// AFK (away from keyboard).
//
// The alert may contain the currently active task.
type AlertFormerlyAFK struct {
	alertUnion
	ActiveTask *Task
	AFKSince   time.Time
}

func (AlertFormerlyAFK) isAlertUnion() {}
