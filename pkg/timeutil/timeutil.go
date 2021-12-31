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

// Package timeutil contains some types and functions to help work with times
// and time spans.
package timeutil

import (
	"fmt"
	"time"
)

// TimeSpan holds a start and end timestamp. Both the start and the end are
// optional, as represented by being set to nil.
type TimeSpan struct {
	Start *time.Time
	End   *time.Time
}

// TimeSpanShorthand is an enumeration of different TimeSpan templates.
type TimeSpanShorthand byte

const (
	// TimeSpanNone represents a TimeSpan of nil - nil
	TimeSpanNone TimeSpanShorthand = iota
	// TimeSpanPast represents a TimeSpan of nil - now
	TimeSpanPast
	// TimeSpanFuture represents a TimeSpan of now - nil
	TimeSpanFuture
	// TimeSpanThisDay represents a TimeSpan of 00:00 today - 23:59 today
	TimeSpanThisDay
	// TimeSpanThisWeek represents a TimeSpan of 00:00 this monday - 23:59 this sunday
	TimeSpanThisWeek
	// TimeSpanPrevDay represents a TimeSpan of 00:00 yesterday - 23:59 yesterday
	TimeSpanPrevDay
	// TimeSpanPrevWeek represents a TimeSpan of 00:00 monday - 23:59 sunday last week
	TimeSpanPrevWeek
	// TimeSpanNextDay represents a TimeSpan of 00:00 tomorrow - 23:59 tomorrow
	TimeSpanNextDay
	// TimeSpanNextWeek represents a TimeSpan of 00:00 monday - 23:59 sunday next week
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

// Span returns a TimeSpan given a specific reference time of when "now" is.
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

// Day returns a TimeSpan from 00:00 - 23:59 for the same day as "now".
func Day(now time.Time) TimeSpan {
	var (
		y, m, d = now.Date()
		loc     = now.Location()
		start   = time.Date(y, m, d, 0, 0, 0, 0, loc)
		end     = time.Date(y, m, d, 23, 59, 59, 9999, loc)
	)
	return TimeSpan{&start, &end}
}

// Week returns a TimeSpan from 00:00 monday - 23:59 sunday for the same week
// as "now".
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

// DaysSinceMonday returns the number of days has passed since last time it was
// a monday.
//
// This function is a naïve implementation that assumes all weeks have the
// 7 weekdays. In reality, there are some weird edge cases where some regions
// have skipped some days, but those cases are left as "undefined behavior".
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
