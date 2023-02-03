// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package dinkur

import "time"

// TimeFields contains time metadata fields used by multiple other models.
type TimeFields struct {
	// CreatedAt is when the object was created.
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt" xml:"CreatedAt"`
	// UpdatedAt stores when the object was last updated/edited.
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt" xml:"UpdatedAt"`
}

// CommonFields contains fields used by multiple other models.
type CommonFields struct {
	// ID is a unique identifier for this entry. The same ID will never be used
	// twice for a given database.
	ID uint `json:"id" yaml:"id" xml:"Id"`
	TimeFields
}

// Entry is a time tracked entry.
type Entry struct {
	CommonFields `yaml:",inline"`
	// Name of the entry.
	Name string `json:"name" yaml:"name" xml:"Name"`
	// Start time of the entry.
	Start time.Time `json:"start" yaml:"start" xml:"Start"`
	// End time of the entry, or nil if the entry is still active.
	End *time.Time `json:"end" yaml:"end" xml:"End"`
}

// Elapsed returns the duration of the entry. If the entry is currently active,
// the duration is calculated from the start to now.
func (t Entry) Elapsed() time.Duration {
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
	// EventUnknown means the remove Dinkur daemon or client sent an undefined
	// event type.
	EventUnknown EventType = iota
	// EventCreated means the subject was just created.
	EventCreated
	// EventUpdated means the subject was just updated.
	EventUpdated
	// EventDeleted means the subject was just deleted.
	EventDeleted
)

func (ev EventType) String() string {
	switch ev {
	case EventCreated:
		return "created"
	case EventUpdated:
		return "updated"
	case EventDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// Status holds data about the user's status, such as if they're currently AFK.
type Status struct {
	TimeFields
	AFKSince  *time.Time // set if currently AFK
	BackSince *time.Time // set if returned from being AFK
}
