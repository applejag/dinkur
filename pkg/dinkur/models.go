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

// CommonFields contains fields used by multiple other models.
type CommonFields struct {
	// ID is a unique identifier for this entry. The same ID will never be used
	// twice for a given database.
	ID uint `json:"id" yaml:"id" xml:"Id"`
	// CreatedAt is when the object was created.
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt" xml:"CreatedAt"`
	// UpdatedAt stores when the object was last updated/edited.
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt" xml:"UpdatedAt"`
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

type AlertInterface interface {
	isAlertUnion()
	Type() AlertType
}

// Alert defines unexported interface used to restrict the alert union types.
type Alert interface {
	AlertInterface
	Common() CommonFields
}

// NewAlert defines a new alert to be created. The ID and other common fields
// are ignored as they will be set on creation.
type NewAlert interface {
	AlertInterface
}

// EditAlert defines an alert to be updated. The ID is used to identify the
// alert, but the other common fields are ignored as they will be automatically
// updated.
type EditAlert interface {
	AlertInterface
	ID() uint
}

// AlertType is an enum of the different alert types.
type AlertType uint8

const (
	// AlertTypeUnspecified is the default value for the alert type enum.
	// It does not represent any alert type.
	AlertTypeUnspecified AlertType = iota
	// AlertTypePlainMessage represents the AlertPlainMessage type.
	AlertTypePlainMessage
	// AlertTypeAFK represents the AlertAFK type.
	AlertTypeAFK
)

// AlertPlainMessage is a type of alert for generic messages that needs to be
// presented to the user with no need for user action.
type AlertPlainMessage struct {
	CommonFields
	Message string
}

func (AlertPlainMessage) isAlertUnion() {}

// Type returns the enum value of this alert type.
func (AlertPlainMessage) Type() AlertType { return AlertTypePlainMessage }

// Common returns the common model fields: ID, CreatedAt, and UpdatedAt.
func (a AlertPlainMessage) Common() CommonFields { return a.CommonFields }

// AlertAFK is a type of alert that's issued when the user has just become AFK
// (away from keyboard) and when they have returned, both when also having an
// active entry. I.e. no AFK alert is issued when not tracking any entry.
type AlertAFK struct {
	CommonFields
	ActiveEntry Entry
	AFKSince    time.Time
	BackSince   *time.Time
}

func (AlertAFK) isAlertUnion() {}

// Type returns the enum value of this alert type.
func (AlertAFK) Type() AlertType { return AlertTypeAFK }

// Common returns the common model fields: ID, CreatedAt, and UpdatedAt.
func (a AlertAFK) Common() CommonFields { return a.CommonFields }
