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

package pflagutil

import (
	"time"

	"github.com/dinkur/dinkur/internal/fuzzytime"
	"gopkg.in/typ.v4"
)

// TimeDefaultLayout is the layout used when showing a Time flag value in the
// program's helper text.
var TimeDefaultLayout = "Jan 02 15:04"

// Time is a pflag.Value-compatible type for allowing datetimes to be used in
// flags. The fuzzytime package is used to parse the user-provided flag
// string value.
type Time struct {
	Now    bool
	source string
	valid  bool
	parsed time.Time
}

// String returns a formatted string of the underlying time. If the Now field
// has been set, the string literal "now" is returned instead.
func (t *Time) String() string {
	if t == nil {
		return ""
	}
	if t.Now {
		return "now"
	}
	if !t.valid {
		return ""
	}
	return t.parsed.Format(TimeDefaultLayout)
}

// Set attempts to parse the string as a time.Time and updates its internal
// state on success, or returns a parsing error if it fails.
func (t *Time) Set(s string) error {
	parsed, err := fuzzytime.Parse(s, time.Now())
	if err != nil {
		return err
	}
	t.valid = true
	t.parsed = parsed
	t.source = s
	t.Now = false
	return nil
}

// Type returns "time", the flag type name to be used in helper text.
func (t *Time) Type() string {
	return "time"
}

// Time returns the time.Time value. If the Now field is set, time.Now() is
// returned instead.
func (t *Time) Time(base time.Time) time.Time {
	if t.Now || !t.valid {
		return time.Now()
	}
	parsed, _ := fuzzytime.Parse(t.source, base)
	return parsed
}

// TimePtr returns the time.Time value, or nil if the object is nil. If the Now
// field is set, time.Now() is returned instead.
func (t *Time) TimePtr(base time.Time) *time.Time {
	if t == nil {
		return nil
	}
	if t.Now {
		return typ.Ref(time.Now())
	}
	if !t.valid {
		return nil
	}
	parsed, _ := fuzzytime.Parse(t.source, base)
	return &parsed
}

// Source returns the source string that was used to create this time.
func (t *Time) Source() string {
	if t == nil {
		return ""
	}
	if t.Now {
		return "now"
	}
	if !t.valid {
		return ""
	}
	return t.source
}
