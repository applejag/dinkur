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

package togrpc

import (
	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/timeutil"
)

// Shorthand converts a timeutil shorthand to a gRPC shorthand.
func Shorthand(s timeutil.TimeSpanShorthand) dinkurapiv1.GetEntryListRequest_Shorthand {
	switch s {
	case timeutil.TimeSpanPast:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_PAST
	case timeutil.TimeSpanFuture:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_FUTURE
	case timeutil.TimeSpanThisDay:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_THIS_DAY
	case timeutil.TimeSpanThisWeek:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_THIS_MON_TO_SUN
	case timeutil.TimeSpanPrevDay:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_PREV_DAY
	case timeutil.TimeSpanPrevWeek:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_PREV_MON_TO_SUN
	case timeutil.TimeSpanNextDay:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_NEXT_DAY
	case timeutil.TimeSpanNextWeek:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_NEXT_MON_TO_SUN
	default:
		return dinkurapiv1.GetEntryListRequest_SHORTHAND_UNSPECIFIED
	}
}
