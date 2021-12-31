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

package pflagutil

import (
	"time"

	"github.com/dinkur/dinkur/internal/fuzzytime"
)

var TimeDefaultLayout = "Jan 02 15:04"

type Time struct {
	Now  bool
	time time.Time
}

func (t *Time) String() string {
	if t == nil {
		return ""
	}
	if t.Now {
		return "now"
	}
	return time.Time(t.time).Format(TimeDefaultLayout)
}

func (t *Time) Set(s string) error {
	parsed, err := fuzzytime.Parse(s)
	if err != nil {
		return err
	}
	t.time = parsed
	t.Now = false
	return nil
}

func (t *Time) Type() string {
	return "time"
}

func (t *Time) Time() time.Time {
	if t.Now {
		return time.Now()
	}
	return time.Time(t.time)
}

func (t *Time) TimePtr() *time.Time {
	if t == nil {
		return nil
	}
	if t.Now {
		now := time.Now()
		return &now
	}
	return (*time.Time)(&t.time)
}
