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
	"github.com/dinkur/dinkur/pkg/fromgrpc"
	"github.com/dinkur/dinkur/pkg/togrpc"
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
			alert, err := fromgrpc.AlertPtr(res.Alert)
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
				Alert: alert,
				Event: fromgrpc.Event(res.Event),
			}
		}
	}()
	return alertChan, nil
}

func (c *client) CreateAlert(ctx context.Context, newAlert dinkur.NewAlert) (dinkur.Alert, error) {
	res, err := invoke(ctx, c, c.alerter.CreateAlert, &dinkurapiv1.CreateAlertRequest{
		Type: togrpc.AlertData(newAlert.Alert),
	})
	if err != nil {
		return nil, convError(err)
	}
	alert, err := fromgrpc.AlertPtrNoNil(res.Alert)
	if err != nil {
		return nil, convError(err)
	}
	return alert, nil
}

func (c *client) CreateOrUpdateAlertByType(ctx context.Context, newAlert dinkur.NewAlert) (dinkur.NewOrUpdatedAlert, error) {
	res, err := invoke(ctx, c, c.alerter.CreateOrUpdateAlert, &dinkurapiv1.CreateOrUpdateAlertRequest{
		Type: togrpc.AlertData(newAlert.Alert),
	})
	if err != nil {
		return dinkur.NewOrUpdatedAlert{}, convError(err)
	}
	alertBefore, err := fromgrpc.AlertPtr(res.Before)
	if err != nil {
		return dinkur.NewOrUpdatedAlert{}, convError(err)
	}
	alertAfter, err := fromgrpc.AlertPtrNoNil(res.After)
	if err != nil {
		return dinkur.NewOrUpdatedAlert{}, convError(err)
	}
	return dinkur.NewOrUpdatedAlert{
		Before: alertBefore,
		After:  alertAfter,
	}, nil
}

func (c *client) GetAlertList(ctx context.Context) ([]dinkur.Alert, error) {
	res, err := invoke(ctx, c, c.alerter.GetAlertList, &dinkurapiv1.GetAlertListRequest{})
	if err != nil {
		return nil, convError(err)
	}
	alerts, err := fromgrpc.AlertSlice(res.Alerts)
	if err != nil {
		return nil, convError(err)
	}
	return alerts, nil
}

func (c *client) UpdateAlert(ctx context.Context, edit dinkur.EditAlert) (dinkur.UpdatedAlert, error) {
	res, err := invoke(ctx, c, c.alerter.UpdateAlert, &dinkurapiv1.UpdateAlertRequest{
		Id:   uint64(edit.ID),
		Type: togrpc.AlertData(edit.Alert),
	})
	if err != nil {
		return dinkur.UpdatedAlert{}, convError(err)
	}
	alertBefore, err := fromgrpc.AlertPtrNoNil(res.Before)
	if err != nil {
		return dinkur.UpdatedAlert{}, convError(err)
	}
	alertAfter, err := fromgrpc.AlertPtrNoNil(res.After)
	if err != nil {
		return dinkur.UpdatedAlert{}, convError(err)
	}
	return dinkur.UpdatedAlert{
		Before: alertBefore,
		After:  alertAfter,
	}, nil
}

func (c *client) DeleteAlert(ctx context.Context, id uint) (dinkur.Alert, error) {
	res, err := invoke(ctx, c, c.alerter.DeleteAlert, &dinkurapiv1.DeleteAlertRequest{
		Id: uint64(id),
	})
	if err != nil {
		return nil, convError(err)
	}
	alert, err := fromgrpc.AlertPtrNoNil(res.DeletedAlert)
	if err != nil {
		return nil, convError(err)
	}
	return alert, nil
}

func (c *client) DeleteAlertByType(ctx context.Context, alertType dinkur.AlertType) (dinkur.Alert, error) {
	res, err := invoke(ctx, c, c.alerter.DeleteAlertType, &dinkurapiv1.DeleteAlertTypeRequest{
		Type: togrpc.AlertType(alertType),
	})
	if err != nil {
		return nil, convError(err)
	}
	alert, err := fromgrpc.AlertPtrNoNil(res.DeletedAlert)
	if err != nil {
		return nil, convError(err)
	}
	return alert, nil
}
