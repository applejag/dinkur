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

// Event converts a Go event type to a gRPC event type.
func Event(ev dinkur.EventType) dinkurapiv1.Event {
	switch ev {
	case dinkur.EventCreated:
		return dinkurapiv1.Event_EVENT_CREATED
	case dinkur.EventUpdated:
		return dinkurapiv1.Event_EVENT_UPDATED
	case dinkur.EventDeleted:
		return dinkurapiv1.Event_EVENT_DELETED
	default:
		return dinkurapiv1.Event_EVENT_UNSPECIFIED
	}
}
