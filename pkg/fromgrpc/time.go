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

package fromgrpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/typ.v4"
)

// TimePtr converts gRPC timestamp to Go time pointer.
func TimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	return typ.Ref(ts.AsTime().Local())
}

// TimeOrZero converts gRPC timestamp to Go time, or zero if nil.
func TimeOrZero(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().Local()
}

// TimeOrNow converts gRPC timestamp to Go time, or time.Now() if nil.
func TimeOrNow(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Now()
	}
	return ts.AsTime()
}
