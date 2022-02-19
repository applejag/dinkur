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
	ErrAlreadyConnected    = errors.New("client is already connected to database")
	ErrNotConnected        = errors.New("client is not connected to database")
	ErrEntryNameEmpty      = errors.New("entry name cannot be empty")
	ErrEntryEndBeforeStart = errors.New("entry end time cannot be before start time")
	ErrNotFound            = gorm.ErrRecordNotFound
	ErrLimitTooLarge       = errors.New("search limit is too large, maximum: " + strconv.Itoa(math.MaxInt))
	ErrClientIsNil         = errors.New("client is nil")
)

// Client is a Dinkur client interface. This is the core interface to act upon
// the Dinkur data store. Depending on the implementation, it may either talk
// directly to an Sqlite3 database file, or talk to a Dinkur daemon via gRPC
// over TCP/IP that in turn talks to a database.
type Client interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	Entries
	Alerter
}

// Entries is the Dinkur client methods targeted to reading, creating, and
// updating entries.
type Entries interface {
	GetEntry(ctx context.Context, id uint) (Entry, error)
	GetEntryList(ctx context.Context, search SearchEntry) ([]Entry, error)
	GetActiveEntry(ctx context.Context) (*Entry, error)
	UpdateEntry(ctx context.Context, edit EditEntry) (UpdatedEntry, error)
	DeleteEntry(ctx context.Context, id uint) (Entry, error)
	CreateEntry(ctx context.Context, entry NewEntry) (StartedEntry, error)
	StopActiveEntry(ctx context.Context, endTime time.Time) (*Entry, error)
	StreamEntry(ctx context.Context) (<-chan StreamedEntry, error)
}

// Alerter is the Dinkur client methods targeted to reading alerts.
type Alerter interface {
	StreamAlert(ctx context.Context) (<-chan StreamedAlert, error)
	CreateAlert(ctx context.Context, newAlert NewAlert) (Alert, error)
	CreateOrUpdateAlertByType(ctx context.Context, newAlert NewAlert) (NewOrUpdatedAlert, error)
	GetAlertList(ctx context.Context) ([]Alert, error)
	UpdateAlert(ctx context.Context, edit EditAlert) (UpdatedAlert, error)
	DeleteAlert(ctx context.Context, id uint) (Alert, error)
	DeleteAlertByType(ctx context.Context, alertType AlertType) (Alert, error)
}

// StreamedAlert holds an alert and its event type.
type StreamedAlert struct {
	Alert Alert
	Event EventType
}

// SearchEntry holds parameters used when searching for list of entries.
type SearchEntry struct {
	Start *time.Time
	End   *time.Time
	Limit uint

	Shorthand          timeutil.TimeSpanShorthand
	NameFuzzy          string
	NameHighlightStart string
	NameHighlightEnd   string
}

// EditEntry holds parameters used when editing a entry.
type EditEntry struct {
	// IDOrZero of the entry to edit. If set to nil, then Dinkur will attempt to make
	// an educated guess on what entry to edit by editing the active entry or a
	// recent entry.
	IDOrZero uint
	// Name is the new entry name. If AppendName is enabled, then this value will
	// append to the existing name, delimited with a space.
	//
	// No change to the entry name is applied if this is set to nil.
	Name *string
	// Start is the new entry start timestamp.
	//
	// No change to the entry start timestamp is applied if this is set to nil.
	Start *time.Time
	// End is the new entry end timestamp.
	//
	// No change to the entry end timestamp is applied if this is set to nil.
	End *time.Time
	// AppendName changes the name field to append the name to the entry's
	// existing name (delimited with a space) instead of replacing it.
	AppendName         bool
	StartAfterIDOrZero uint
	EndBeforeIDOrZero  uint
	StartAfterLast     bool
}

// UpdatedEntry is the response from an edited entry, with values for before the
// edits were applied and after they were applied.
type UpdatedEntry struct {
	Before Entry
	After  Entry
}

// NewEntry holds parameters used when creating a new entry.
type NewEntry struct {
	Name               string
	Start              *time.Time
	End                *time.Time
	StartAfterIDOrZero uint
	EndBeforeIDOrZero  uint
	StartAfterLast     bool
}

// StartedEntry is the response from creating a new entry, with the newly created
// entry object as well as the entry that was stopped when creating the entry,
// if any entry was previously active.
type StartedEntry struct {
	Started Entry
	Stopped *Entry
}

// StreamedEntry holds a entry and its event type.
type StreamedEntry struct {
	Entry Entry
	Event EventType
}

// NewOrUpdatedAlert is the response from a create-or-update call. The Before
// field can be nil if the operation resulted in a created alert.
type NewOrUpdatedAlert struct {
	Before Alert // will be nil if alert was created
	After  Alert
}

// UpdatedAlert is the response from an edited alert, with values for before the
// edits were applied and after they were applied.
type UpdatedAlert struct {
	Before Alert
	After  Alert
}
