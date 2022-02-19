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
	"gopkg.in/typ.v2"
)

// Alert converts a Go alert to a gRPC alert.
func Alert(alert dinkur.Alert) *dinkurapiv1.Alert {
	common := alert.Common()
	a := &dinkurapiv1.Alert{
		Id:      uint64(common.ID),
		Created: Timestamp(common.CreatedAt),
		Updated: Timestamp(common.UpdatedAt),
	}
	switch alertType := alert.(type) {
	case dinkur.AlertPlainMessage:
		a.Type = &dinkurapiv1.AlertType{
			Data: &dinkurapiv1.AlertType_PlainMessage{
				PlainMessage: AlertPlainMessage(alertType),
			},
		}
	case dinkur.AlertAFK:
		a.Type = &dinkurapiv1.AlertType{
			Data: &dinkurapiv1.AlertType_Afk{
				Afk: AlertAFK(alertType),
			},
		}
	}
	return a
}

// AlertPlainMessage converts a Go plain message alert to a gRPC alert.
func AlertPlainMessage(alert dinkur.AlertPlainMessage) *dinkurapiv1.AlertPlainMessage {
	return &dinkurapiv1.AlertPlainMessage{
		Message: alert.Message,
	}
}

// AlertAFK converts a Go AFK alert to a gRPC alert.
func AlertAFK(alert dinkur.AlertAFK) *dinkurapiv1.AlertAfk {
	return &dinkurapiv1.AlertAfk{
		ActiveEntry: EntryPtr(&alert.ActiveEntry),
		AfkSince:    Timestamp(alert.AFKSince),
		BackSince:   TimestampPtr(alert.BackSince),
	}
}

// AlertSlice converts a slice of Go alerts to gRPC alerts.
func AlertSlice(slice []dinkur.Alert) []*dinkurapiv1.Alert {
	return typ.Map(slice, Alert)
}
