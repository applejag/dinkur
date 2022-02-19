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
	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Status converts a dinkur status to a gRPC status.
func Status(status dinkur.Status) *dinkurapiv1.Status {
	return &dinkurapiv1.Status{
		Created:   Timestamp(status.CreatedAt),
		Updated:   Timestamp(status.UpdatedAt),
		AfkSince:  TimestampPtr(status.AFKSince),
		BackSince: TimestampPtr(status.BackSince),
	}
}
