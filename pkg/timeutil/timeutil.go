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

package timeutil

import "time"

type TimeSpan struct {
	Start time.Time
	End   time.Time
}

type TimeSpanShorthand byte

const (
	TimeSpanNone TimeSpanShorthand = iota
	TimeSpanThisDay
	TimeSpanThisWeek
)

func (s TimeSpanShorthand) Span(now time.Time) TimeSpan {
	switch s {
	case TimeSpanThisDay:
		return Today(now)
	case TimeSpanThisWeek:
		return Week(now)
	default:
		return TimeSpan{now, now}
	}
}

func Today(now time.Time) TimeSpan {
	var (
		y, m, d = now.Date()
		loc     = now.Location()
	)
	return TimeSpan{
		Start: time.Date(y, m, d, 0, 0, 0, 0, loc),
		End:   time.Date(y, m, d, 23, 59, 59, 9999, loc),
	}
}

func Week(now time.Time) TimeSpan {
	var (
		y, m, d     = now.Date()
		loc         = now.Location()
		wd          = now.Weekday()
		sinceMonday = DaysSinceMonday(wd)
	)
	return TimeSpan{
		Start: time.Date(y, m, d-sinceMonday, 0, 0, 0, 0, loc),
		End:   time.Date(y, m, d+6-sinceMonday, 23, 59, 59, 9999, loc),
	}
}

func DaysSinceMonday(day time.Weekday) int {
	switch day {
	case time.Tuesday:
		return 1
	case time.Wednesday:
		return 2
	case time.Thursday:
		return 3
	case time.Friday:
		return 4
	case time.Saturday:
		return 5
	case time.Sunday:
		return 6
	default:
		return 0
	}
}
