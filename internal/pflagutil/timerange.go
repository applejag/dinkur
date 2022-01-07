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
	"fmt"
	"strings"

	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/spf13/cobra"
)

// NewTimeRangePtr returns a pointer to a new TimeRange instance.
func NewTimeRangePtr(shorthand timeutil.TimeSpanShorthand) *TimeRange {
	return (*TimeRange)(&shorthand)
}

// TimeRange is a pflag.Value-compatible type for allowing time span shorthand
// enumeration to be used in flags.
type TimeRange timeutil.TimeSpanShorthand

// String returns the string representation of the time range.
func (r *TimeRange) String() string {
	if r == nil {
		return ""
	}
	switch timeutil.TimeSpanShorthand(*r) {
	case timeutil.TimeSpanNone:
		return "all"
	case timeutil.TimeSpanPast:
		return "past"
	case timeutil.TimeSpanFuture:
		return "future"
	case timeutil.TimeSpanThisDay:
		return "today"
	case timeutil.TimeSpanThisWeek:
		return "week"
	case timeutil.TimeSpanPrevDay:
		return "yesterday"
	case timeutil.TimeSpanPrevWeek:
		return "lastweek"
	case timeutil.TimeSpanNextDay:
		return "tomorrow"
	case timeutil.TimeSpanNextWeek:
		return "nextweek"
	default:
		return ""
	}
}

// Set attempts to parse the string as a timeutil.TimeSpanShorthand and updates
// its internal state on success, or returns a parsing error if it fails.
func (r *TimeRange) Set(s string) error {
	parsed, ok := parseShorthand(s)
	if !ok {
		return fmt.Errorf("unknown time range: %q", s)
	}
	*r = TimeRange(parsed)
	return nil
}

// Type returns "range", the flag type name to be used in helper text.
func (r *TimeRange) Type() string {
	return "range"
}

// TimeSpanShorthand returns the underlying timeutil value.
func (r *TimeRange) TimeSpanShorthand() timeutil.TimeSpanShorthand {
	if r == nil {
		return timeutil.TimeSpanNone
	}
	return timeutil.TimeSpanShorthand(*r)
}

func parseShorthand(s string) (timeutil.TimeSpanShorthand, bool) {
	switch strings.ToLower(s) {
	case "all", "a":
		return timeutil.TimeSpanNone, true
	case "past", "p":
		return timeutil.TimeSpanPast, true
	case "future", "f":
		return timeutil.TimeSpanFuture, true
	case "today", "t":
		return timeutil.TimeSpanThisDay, true
	case "week", "w":
		return timeutil.TimeSpanThisWeek, true
	case "yesterday", "y", "ld":
		return timeutil.TimeSpanPrevDay, true
	case "lastweek", "lw":
		return timeutil.TimeSpanPrevWeek, true
	case "tomorrow", "nd":
		return timeutil.TimeSpanNextDay, true
	case "nextweek", "nw":
		return timeutil.TimeSpanNextWeek, true
	default:
		return timeutil.TimeSpanNone, false
	}
}

// TimeRangeCompletion is a completion function for the TimeRange flag type, and
// is meant to be registered into a Cobra command object.
func TimeRangeCompletion(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"all\tlist all tasks, i.e. no baseline",
		"a\tlist all tasks, i.e. no baseline",
		"past\tlist all tasks before now",
		"p\tlist all tasks before now",
		"future\tlist all tasks since now",
		"f\tlist all tasks since now",
		"today\tlist today's tasks (default)",
		"t\tlist today's tasks (default)",
		"week\tlist this week's tasks",
		"w\tlist this week's tasks",
		"yesterday\tlist yesterday's tasks",
		"y\tlist yesterday's tasks",
		"ld\tlist yesterday's tasks",
		"lastweek\tlist last week's tasks",
		"lw\tlist last week's tasks",
		"tomorrow\tlist tomorrow's tasks",
		"nd\tlist tomorrow's tasks",
		"nextweek\tlist next week's tasks",
		"nw\tlist next week's tasks",
	}, cobra.ShellCompDirectiveDefault
}
