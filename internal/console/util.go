// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
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

package console

import (
	"fmt"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

func newDate(year int, month time.Month, day int) date {
	return date{year, month, day}
}

type date struct {
	year  int
	month time.Month
	day   int
}

func (d date) String() string {
	return fmt.Sprintf("%s %d", d.month.String()[:3], d.day)
}

type taskDateGroup struct {
	date  date
	tasks []dinkur.Task
}

// groupTasksByDate assumes the slice is already sorted on task.Start
func groupTasksByDate(tasks []dinkur.Task) []taskDateGroup {
	if len(tasks) == 0 {
		return nil
	}
	var groups []taskDateGroup
	var group taskDateGroup
	for _, t := range tasks {
		d := newDate(t.Start.Date())
		if d != group.date {
			if len(group.tasks) > 0 {
				groups = append(groups, group)
			}
			group = taskDateGroup{date: d}
		}
		group.tasks = append(group.tasks, t)
	}
	if len(group.tasks) > 0 {
		groups = append(groups, group)
	}
	return groups
}

type taskSum struct {
	start    time.Time
	end      *time.Time
	duration time.Duration
}

// sumTasks assumes the slice is already sorted on task.Start
func sumTasks(tasks []dinkur.Task) taskSum {
	if len(tasks) == 0 {
		return taskSum{}
	}
	sum := taskSum{start: tasks[0].Start}
	var anyNilEnd bool
	for _, t := range tasks {
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
