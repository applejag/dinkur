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

package dinkurclient

import (
	"context"
	"io"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

func (c *client) StreamAlert(ctx context.Context) (<-chan dinkur.StreamedAlert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	stream, err := c.alerter.StreamAlert(ctx, &dinkurapiv1.StreamAlertRequest{})
	if err != nil {
		return nil, convError(err)
	}
	alertChan := make(chan dinkur.StreamedAlert)
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Error().
						WithError(convError(err)).
						Message("Error when streaming alerts. Closing stream.")
				}
				close(alertChan)
				return
			}
			if res == nil {
				continue
			}
			const logWarnMsg = "Error when streaming alerts. Ignoring message."
			alert, err := convAlertPtr(res.Alert)
			if err != nil {
				log.Warn().WithError(convError(err)).
					Message(logWarnMsg)
				continue
			}
			if alert == nil {
				log.Warn().WithError(ErrUnexpectedNilAlert).
					Message(logWarnMsg)
				continue
			}
			alertChan <- dinkur.StreamedAlert{
				Alert: *alert,
				Event: convEvent(res.Event),
			}
		}
	}()
	return alertChan, nil
}

func (c *client) GetAlertList(ctx context.Context) ([]dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.alerter.GetAlertList(ctx, &dinkurapiv1.GetAlertListRequest{})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	alerts, err := convAlertSlice(res.Alerts)
	if err != nil {
		return nil, convError(err)
	}
	return alerts, nil
}

func (c *client) DeleteAlert(ctx context.Context, id uint) (dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.alerter.DeleteAlert(ctx, &dinkurapiv1.DeleteAlertRequest{
		Id: uint64(id),
	})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	alert, err := convAlertPtr(res.DeletedAlert)
	if err != nil {
		return nil, convError(err)
	}
	if alert == nil {
		return nil, ErrUnexpectedNilAlert
	}
	return *alert, nil
}
