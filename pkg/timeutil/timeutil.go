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

import (
	"fmt"
	"time"
)

type TimeSpan struct {
	Start *time.Time
	End   *time.Time
}

type TimeSpanShorthand byte

const (
	TimeSpanNone TimeSpanShorthand = iota
	TimeSpanPast
	TimeSpanFuture
	TimeSpanThisDay
	TimeSpanThisWeek
	TimeSpanPrevDay
	TimeSpanPrevWeek
	TimeSpanNextDay
	TimeSpanNextWeek
)

func (s TimeSpanShorthand) String() string {
	switch s {
	case TimeSpanNone:
		return "none"
	case TimeSpanPast:
		return "past"
	case TimeSpanFuture:
		return "future"
	case TimeSpanThisDay:
		return "day"
	case TimeSpanThisWeek:
		return "week"
	case TimeSpanPrevDay:
		return "yesterday"
	case TimeSpanPrevWeek:
		return "last week"
	case TimeSpanNextDay:
		return "tomorrow"
	case TimeSpanNextWeek:
		return "next week"
	default:
		return fmt.Sprintf("%[1]T(%[1]d)", s)
	}
}

func (s TimeSpanShorthand) Span(now time.Time) TimeSpan {
	switch s {
	case TimeSpanPast:
		return TimeSpan{nil, &now}
	case TimeSpanFuture:
		return TimeSpan{&now, nil}
	case TimeSpanThisDay:
		return Day(now)
	case TimeSpanThisWeek:
		return Week(now)
	case TimeSpanPrevDay:
		return Day(now.Add(-24 * time.Hour))
	case TimeSpanPrevWeek:
		return Week(now.Add(-7 * 24 * time.Hour))
	case TimeSpanNextDay:
		return Day(now.Add(24 * time.Hour))
	case TimeSpanNextWeek:
		return Week(now.Add(7 * 24 * time.Hour))
	default:
		return TimeSpan{}
	}
}

func Day(now time.Time) TimeSpan {
	var (
		y, m, d = now.Date()
		loc     = now.Location()
		start   = time.Date(y, m, d, 0, 0, 0, 0, loc)
		end     = time.Date(y, m, d, 23, 59, 59, 9999, loc)
	)
	return TimeSpan{&start, &end}
}

func Week(now time.Time) TimeSpan {
	var (
		y, m, d     = now.Date()
		loc         = now.Location()
		wd          = now.Weekday()
		sinceMonday = DaysSinceMonday(wd)
		start       = time.Date(y, m, d-sinceMonday, 0, 0, 0, 0, loc)
		end         = time.Date(y, m, d+6-sinceMonday, 23, 59, 59, 9999, loc)
	)
	return TimeSpan{&start, &end}
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
