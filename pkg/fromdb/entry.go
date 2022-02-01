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

package fromdb

import (
	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"gopkg.in/typ.v1"
)

// Entry converts a dbmodel entry to a dinkur entry.
func Entry(t dbmodel.Entry) dinkur.Entry {
	return dinkur.Entry{
		CommonFields: CommonFields(t.CommonFields),
		Name:         t.Name,
		Start:        t.Start.Local(),
		End:          conv.TimePtrLocal(t.End),
	}
}

// EntryPtr converts a dbmodel entry pointer to a dinkur entry, or nil.
func EntryPtr(t *dbmodel.Entry) *dinkur.Entry {
	if t == nil {
		return nil
	}
	return typ.Ptr(Entry(*t))
}

// EntrySlice converts a slice of dbmodel entries to dinkur entries.
func EntrySlice(entries []dbmodel.Entry) []dinkur.Entry {
	return typ.Map(entries, Entry)
}
