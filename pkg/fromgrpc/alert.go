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
	"fmt"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

// AlertPtr converts a gRPC alert to a Go alert.
func AlertPtr(alert *dinkurapiv1.Alert) (*dinkur.Alert, error) {
	if alert == nil {
		return nil, nil
	}
	id, err := conv.Uint64ToUint(alert.Id)
	if err != nil {
		return nil, err
	}
	common := dinkur.CommonFields{
		ID:        id,
		CreatedAt: TimeOrZero(alert.Created),
		UpdatedAt: TimeOrZero(alert.Updated),
	}
	var a dinkur.Alert
	switch alertType := alert.Type.Data.(type) {
	case *dinkurapiv1.AlertType_PlainMessage:
		a = AlertPlainMessage(common, alertType.PlainMessage)
	case *dinkurapiv1.AlertType_Afk:
		at, err := AlertAFK(common, alertType.Afk)
		if err != nil {
			return nil, err
		}
		a = at
	}
	return &a, nil
}

// AlertPlainMessage converts a gRPC plain message alert to a Go alert.
func AlertPlainMessage(common dinkur.CommonFields, alert *dinkurapiv1.AlertPlainMessage) dinkur.AlertPlainMessage {
	if alert == nil {
		return dinkur.AlertPlainMessage{CommonFields: common}
	}
	return dinkur.AlertPlainMessage{
		CommonFields: common,
		Message:      alert.Message,
	}
}

// AlertAFK converts a gRPC AFK alert to a Go alert.
func AlertAFK(common dinkur.CommonFields, alert *dinkurapiv1.AlertAfk) (dinkur.AlertAFK, error) {
	if alert == nil {
		return dinkur.AlertAFK{CommonFields: common}, nil
	}
	entry, err := EntryPtrNoNil(alert.ActiveEntry)
	if err != nil {
		return dinkur.AlertAFK{}, err
	}
	return dinkur.AlertAFK{
		CommonFields: common,
		ActiveEntry:  entry,
		AFKSince:     TimeOrZero(alert.AfkSince),
		BackSince:    TimePtr(alert.BackSince),
	}, nil
}

// AlertSlice converts a slice of gRPC alerts to Go alerts.
func AlertSlice(slice []*dinkurapiv1.Alert) ([]dinkur.Alert, error) {
	entries := make([]dinkur.Alert, 0, len(slice))
	for _, a := range slice {
		a2, err := AlertPtr(a)
		if a2 == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("alert #%d: %w", a.Id, err)
		}
		entries = append(entries, *a2)
	}
	return entries, nil
}

// AlertType converts a gRPC alert type enum to a Dinkur alert type enum.
func AlertType(alertType dinkurapiv1.ALERT) dinkur.AlertType {
	switch alertType {
	case dinkurapiv1.ALERT_ALERT_PLAIN_MESSAGE:
		return dinkur.AlertTypePlainMessage
	case dinkurapiv1.ALERT_ALERT_AFK:
		return dinkur.AlertTypeAFK
	default:
		return dinkur.AlertTypeUnspecified
	}
}
