// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

package dinkurdb

import (
	"errors"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"gopkg.in/typ.v1"
)

var (
	commonFieldsColumnID = "id"
)

// CommonFields contains fields used by multiple other models.
type CommonFields struct {
	// ID is a unique identifier for this entry. The same ID will never be used
	// twice for a given database.
	ID uint `gorm:"primaryKey;autoIncrement;type:INTEGER PRIMARY KEY AUTOINCREMENT"`
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
	entryFieldEnd = "End"

	entryColumnID    = "id"
	entryColumnStart = "start"
	entryColumnEnd   = "end"
)

// Entry is a time tracked entry stored in the database.
type Entry struct {
	CommonFields
	// Name of the entry.
	Name string `gorm:"not null;default:''"`
	// Start time of the entry.
	Start time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	// End time of the entry, or nil if the entry is still active.
	End *time.Time `gorm:"index"`
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

const (
	entryFTS5ColumnRowID = "entries_idx.rowid"
	entryFTS5ColumnName  = "entries_idx.name"
)

// EntryFTS5 is used for free-text searching entries.
type EntryFTS5 struct {
	RowID uint   `gorm:"primaryKey;column:rowid"`
	Name  string `gorm:"not null;default:''"`
}

// TableName overrides the table name used by GORM.
func (EntryFTS5) TableName() string {
	return "entries_idx"
}

const (
	alertColumnPlainMessage = "PlainMessage"
	alertColumnAFK          = "AFK"
)

// Alert is the parent alert type for all alert types. Only one of the inner
// alert types are expected to be set. It's considered undefined behaviour to
// assign multiple alert types to an alert, such as assigning both a plain
// message alert and an AFK alert.
type Alert struct {
	CommonFields
	PlainMessage *AlertPlainMessage `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AFK          *AlertAFK          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// AlertPlainMessage is an arbitrary message the user needs to see.
type AlertPlainMessage struct {
	ID      uint
	AlertID uint

	Message string
}

// AlertAFK is an AFK (Away From Keyboard) alert.
type AlertAFK struct {
	ID      uint
	AlertID uint

	ActiveEntry   Entry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ActiveEntryID uint
	AFKSince      time.Time
	BackSince     *time.Time
}

func timePtrUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return typ.Ptr((*t).UTC())
}

func timePtrLocal(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return typ.Ptr((*t).Local())
}

func convEntry(t Entry) dinkur.Entry {
	return dinkur.Entry{
		CommonFields: convCommonFields(t.CommonFields),
		Name:         t.Name,
		Start:        t.Start.Local(),
		End:          timePtrLocal(t.End),
	}
}

func convEntryPtr(t *Entry) *dinkur.Entry {
	if t == nil {
		return nil
	}
	return typ.Ptr(convEntry(*t))
}

func convAlert(alert Alert) (dinkur.Alert, error) {
	if alert.PlainMessage != nil {
		return convAlertPlainMessage(alert, *alert.PlainMessage), nil
	}
	if alert.AFK != nil {
		return convAlertAFK(alert, *alert.AFK), nil
	}
	return nil, errors.New("alert does not have an associated alert type")
}

func convAlertPlainMessage(alert Alert, plain AlertPlainMessage) dinkur.Alert {
	return dinkur.AlertPlainMessage{
		CommonFields: convCommonFields(alert.CommonFields),
		Message:      plain.Message,
	}
}

func convAlertAFK(alert Alert, afk AlertAFK) dinkur.Alert {
	return dinkur.AlertAFK{
		CommonFields: convCommonFields(alert.CommonFields),
		ActiveEntry:  convEntry(afk.ActiveEntry),
		AFKSince:     afk.AFKSince.Local(),
		BackSince:    timePtrLocal(afk.BackSince),
	}
}

// Migration holds the latest migration revision identifier. At most one row of
// this object is expected to be in the database at any given time.
type Migration struct {
	CommonFields
	Version MigrationVersion
}

// MigrationVersion is an enumeration stating how outdated the database schema is.
type MigrationVersion int

// LatestMigrationVersion is an integer revision identifier for what migration
// was last applied to the database. This is stored in the database to quickly
// figure out if new migrations needs to be applied.
const LatestMigrationVersion MigrationVersion = 7

const (
	// MigrationUnknown means that Dinkur was unable to evaluate the database's
	// migration status. Perhaps due to an error.
	MigrationUnknown MigrationVersion = -1
	// MigrationNeverApplied means the database has never been migrated before.
	// In other words, it's a brand new database.
	MigrationNeverApplied MigrationVersion = 0
	// MigrationUpToDate means the database does not need any further migrations
	// applied.
	MigrationUpToDate MigrationVersion = LatestMigrationVersion
)

func (s MigrationVersion) String() string {
	switch s {
	case MigrationUnknown:
		return "unknown"
	case MigrationNeverApplied:
		return "never applied"
	case MigrationUpToDate:
		return "up to date"
	default:
		return "outdated"
	}
}
