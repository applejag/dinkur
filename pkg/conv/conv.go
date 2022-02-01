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

package conv

import (
	"time"

	"gopkg.in/typ.v1"
)

// DerefOrZero will dereference the value, or return the zero value if it's nil.
func DerefOrZero[T any](s *T) T {
	if s == nil {
		return typ.Zero[T]()
	}
	return *s
}

// ZeroAsNil will return a pointer to the value, or nil if the value is zero.
func ZeroAsNil[T comparable](s T) *T {
	if s == typ.Zero[T]() {
		return nil
	}
	return &s
}

// TimePtrUTC converts a time to UTC time, or nil.
func TimePtrUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return typ.Ptr((*t).UTC())
}

// TimePtrLocal converts a time to local time, or nil.
func TimePtrLocal(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return typ.Ptr((*t).Local())
}
