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

package fromdb

import (
	"errors"

	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"gopkg.in/typ.v1"
)

// Alert converts a dbmodel alert to a dinkur alert.
func Alert(alert dbmodel.Alert) (dinkur.Alert, error) {
	if alert.PlainMessage != nil {
		return AlertPlainMessage(alert.ID, *alert.PlainMessage), nil
	}
	if alert.AFK != nil {
		return AlertAFK(alert.ID, *alert.AFK), nil
	}
	return nil, errors.New("alert does not have an associated alert type")
}

// AlertPlainMessage converts a dbmodel plain message alert to a dinkur alert.
func AlertPlainMessage(id uint, alert dbmodel.AlertPlainMessage) dinkur.Alert {
	return dinkur.AlertPlainMessage{
		CommonFields: CommonFieldsID(alert.CommonFields, id),
		Message:      alert.Message,
	}
}

// AlertAFK converts a dbmodel AFK alert to a dinkur alert.
func AlertAFK(id uint, alert dbmodel.AlertAFK) dinkur.Alert {
	return dinkur.AlertAFK{
		CommonFields: CommonFieldsID(alert.CommonFields, id),
		ActiveEntry:  Entry(alert.ActiveEntry),
		AFKSince:     alert.AFKSince.Local(),
		BackSince:    conv.TimePtrLocal(alert.BackSince),
	}
}

// AlertSlice converts a slice of dbmodel alerts to dinkur alerts.
func AlertSlice(alerts []dbmodel.Alert) ([]dinkur.Alert, error) {
	return typ.MapErr(alerts, Alert)
}
