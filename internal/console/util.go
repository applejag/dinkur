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

package console

import (
	"fmt"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// FormatDuration returns a formatted time.Duration in the format of
// h:mm:ss.
func FormatDuration(d time.Duration) string {
	var (
		totalSeconds = int(d.Seconds())
		hours        = totalSeconds / 60 / 60
		minutes      = totalSeconds / 60 % 60
		seconds      = totalSeconds % 60
	)
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

type group interface {
	fmt.Stringer

	new(_g groupable) group

	check(t time.Time) bool
	count() int
	addEntry(entry dinkur.Entry)
	getEntries() []dinkur.Entry
	getGroupBy() groupable
}

func newDate(year int, month time.Month, day int) date {
	return date{year, month, day}
}

type date struct {
	year  int
	month time.Month
	day   int
}

func (d date) String() string {
	return fmt.Sprintf("%s-%d", d.month.String()[:3], d.day)
}

func (day) new(t time.Time) groupable {
	return day{t.Day(), int(t.Weekday())}
}

type day struct {
	day     int
	weekDay int
}

func (d day) String() string {
	return fmt.Sprintf("%02d", d.day)
}

func (week) new(t time.Time) groupable {
	year, _week := t.ISOWeek()
	return week{year, _week}
}

type week struct {
	year int
	week int
}

func (w week) String() string {
	return fmt.Sprintf("%02d", w.week)
}

func (month) new(t time.Time) groupable {
	year, _month := t.Year(), t.Month()
	return month{year, _month}
}

type month struct {
	year  int
	month time.Month
}

func (m month) String() string {
	return fmt.Sprintf("%s", m.month.String()[:3])
}

type groupable interface {
	fmt.Stringer
	new(t time.Time) groupable
}

type entryGroup struct {
	groupBy groupable

	entries []dinkur.Entry
}

func (g *entryGroup) new(_g groupable) group {
	return &entryGroup{
		groupBy: _g,
	}
}

func (g *entryGroup) String() string {
	return g.groupBy.String()
}

func (g *entryGroup) addEntry(entry dinkur.Entry) {
	g.entries = append(g.entries, entry)
}

func (g *entryGroup) check(t time.Time) bool {
	return g.groupBy.new(t) == g.groupBy
}

func (g *entryGroup) count() int {
	return len(g.entries)
}

func (g *entryGroup) getEntries() []dinkur.Entry {
	return g.entries
}

func (g *entryGroup) getGroupBy() groupable {
	return g.groupBy
}

// groupEntries assumes the slice is already sorted on entry.Start
func groupEntries(g group, entries []dinkur.Entry) []group {
	if len(entries) == 0 {
		return nil
	}
	var groups []group
	for _, t := range entries {
		if !g.check(t.Start) {
			if g.count() > 0 {
				groups = append(groups, g)
			}
			g = g.new(g.getGroupBy().new(t.Start))
		}
		g.addEntry(t)
	}
	if g.count() > 0 {
		groups = append(groups, g)
	}
	return groups
}

type entrySum struct {
	start    time.Time
	end      *time.Time
	duration time.Duration
}

// sumEntries assumes the slice is already sorted on entry.Start
func sumEntries(entries []dinkur.Entry) entrySum {
	if len(entries) == 0 {
		return entrySum{}
	}
	sum := entrySum{start: entries[0].Start}
	var anyNilEnd bool
	for _, t := range entries {
		sum.duration += t.Elapsed()
		if t.End == nil {
			anyNilEnd = true
		} else if sum.end == nil || t.End.After(*sum.end) {
			sum.end = t.End
		}
	}
	if anyNilEnd {
		sum.end = nil
	}
	return sum
}

func uintWidth(i uint) int {
	switch {
	case i < 1e1:
		return 1
	case i < 1e2:
		return 2
	case i < 1e3:
		return 3
	case i < 1e4:
		return 4
	case i < 1e5:
		return 5
	case i < 1e6:
		return 6
	case i < 1e7:
		return 7
	case i < 1e8:
		return 8
	case i < 1e9:
		return 9
	case i < 1e10:
		return 10
	case i < 1e11:
		return 11
	case i < 1e12:
		return 12
	case i < 1e13:
		return 13
	case i < 1e14:
		return 14
	case i < 1e15:
		return 15
	case i < 1e16:
		return 16
	case i < 1e17:
		return 17
	case i < 1e18:
		return 18
	case i < 1e19:
		return 19
	default:
		return 20
	}
}

func timesEqual(a, b time.Time) bool {
	return a.UnixMilli() == b.UnixMilli()
}

func timesPtrsEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return timesEqual(*a, *b)
}
