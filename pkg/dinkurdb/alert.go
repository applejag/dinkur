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

package dinkurdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/fromdb"
	"gorm.io/gorm"
)

func (c *client) StreamAlert(ctx context.Context) (<-chan dinkur.StreamedAlert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	ch := make(chan dinkur.StreamedAlert)
	go func() {
		done := ctx.Done()
		dbAlertChan := c.alertObs.Sub()
		defer close(ch)
		defer func() {
			if err := c.alertObs.Unsub(dbAlertChan); err != nil {
				log.Warn().WithError(err).Message("Failed to unsub alert.")
			}
		}()
		for {
			select {
			case ev, ok := <-dbAlertChan:
				if !ok {
					return
				}
				alert, err := fromdb.Alert(ev.dbAlert)
				if err != nil {
					log.Warn().
						WithError(err).
						WithUint("alertId", ev.dbAlert.ID).
						Message("Invalid alert event.")
					continue
				}
				ch <- dinkur.StreamedAlert{
					Alert: alert,
					Event: ev.event,
				}
			case <-done:
				return
			}
		}
	}()
	return ch, nil
}

func (c *client) CreateAlert(ctx context.Context, newAlert dinkur.NewAlert) (dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbAlert, err := c.withContext(ctx).createDBAlertAtom(newAlert)
	if err != nil {
		return nil, err
	}
	return fromdb.Alert(dbAlert)
}

func (c *client) createDBAlertAtom(newAlert dinkur.NewAlert) (dbmodel.Alert, error) {
	var dbAlert dbmodel.Alert
	switch alert := newAlert.(type) {
	case dinkur.AlertAFK:
		dbAlert.AFK = &dbmodel.AlertAFK{
			AFKSince:      alert.AFKSince.UTC(),
			BackSince:     conv.TimePtrUTC(alert.BackSince),
			ActiveEntryID: alert.ActiveEntry.ID,
		}
	case dinkur.AlertPlainMessage:
		dbAlert.PlainMessage = &dbmodel.AlertPlainMessage{
			Message: alert.Message,
		}
	default:
		return dbmodel.Alert{}, fmt.Errorf("unsupported alert type: %v", newAlert.Type())
	}
	if err := c.db.Create(&dbAlert).Error; err != nil {
		return dbmodel.Alert{}, err
	}
	return dbAlert, nil
}

func (c *client) CreateOrUpdateAlertByType(ctx context.Context, newAlert dinkur.NewAlert) (dinkur.NewOrUpdatedAlert, error) {
	return dinkur.NewOrUpdatedAlert{}, errors.New("not implemented")
}

func (c *client) GetAlertList(ctx context.Context) ([]dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbAlerts, err := c.withContext(ctx).listDBAlertsAtom()
	if err != nil {
		return nil, err
	}
	alerts, err := fromdb.AlertSlice(dbAlerts)
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

func (c *client) listDBAlertsAtom() ([]dbmodel.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	var dbAlerts []dbmodel.Alert
	if err := c.dbAlertPreloaded().Find(&dbAlerts).Error; err != nil {
		return nil, err
	}
	return dbAlerts, nil
}

func (c *client) UpdateAlert(ctx context.Context, edit dinkur.EditAlert) (dinkur.Alert, error) {
	return nil, errors.New("not implemented")
}

func (c *client) DeleteAlert(ctx context.Context, id uint) (dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbAlert, err := c.withContext(ctx).deleteDBAlertTran(id)
	if err != nil {
		return nil, err
	}
	c.alertObs.PubWait(alertEvent{
		dbAlert: dbAlert,
		event:   dinkur.EventDeleted,
	})
	return nil, nil
}

func (c *client) deleteDBAlertTran(id uint) (dbmodel.Alert, error) {
	var dbAlert dbmodel.Alert
	err := c.transaction(func(tx *client) (tranErr error) {
		dbAlert, tranErr = tx.deleteDBAlertNoTran(id)
		return
	})
	return dbAlert, err
}

func (c *client) DeleteAlertByType(ctx context.Context, alertType dinkur.AlertType) (dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbAlert, err := c.withContext(ctx).deleteDBAlertByType(alertType)
	if err != nil {
		return nil, err
	}
	c.alertObs.PubWait(alertEvent{
		dbAlert: dbAlert,
		event:   dinkur.EventDeleted,
	})
	return fromdb.Alert(dbAlert)
}

func (c *client) deleteDBAlertByType(alertType dinkur.AlertType) (dbmodel.Alert, error) {
	var dbAlert dbmodel.Alert
	err := c.transaction(func(tx *client) (tranErr error) {
		dbAlert, tranErr = tx.deleteDBAlertByTypeNoTran(alertType)
		return
	})
	return dbAlert, err
}

func (c *client) deleteDBAlertByTypeNoTran(alertType dinkur.AlertType) (dbmodel.Alert, error) {
	dbAlert, err := c.getDBAlertByTypeNoTran(alertType)
	if err != nil {
		return dbmodel.Alert{}, fmt.Errorf("get alert by type to delete: %w", err)
	}
	if err := c.db.Delete(&dbmodel.Alert{}, dbAlert.ID).Error; err != nil {
		return dbmodel.Alert{}, fmt.Errorf("delete alert: %w", err)
	}
	return dbAlert, nil
}

func (c *client) getDBAlertByTypeNoTran(alertType dinkur.AlertType) (dbmodel.Alert, error) {
	var joinField string
	switch alertType {
	case dinkur.AlertTypePlainMessage:
		joinField = dbmodel.AlertFieldPlainMessage
	case dinkur.AlertTypeAFK:
		joinField = dbmodel.AlertFieldAFK
	default:
		return dbmodel.Alert{}, fmt.Errorf("unsupported alert type: %v", alertType)
	}
	var dbAlert dbmodel.Alert
	err := c.db.
		Joins(joinField).
		First(&dbAlert).
		Error
	if err != nil {
		return dbmodel.Alert{}, err
	}
	return dbAlert, nil
}

func (c *client) deleteDBAlertNoTran(id uint) (dbmodel.Alert, error) {
	dbAlert, err := c.getDBAlertAtom(id)
	if err != nil {
		return dbmodel.Alert{}, fmt.Errorf("get alert by ID to delete: %w", err)
	}
	if err := c.db.Delete(&dbmodel.Alert{}, id).Error; err != nil {
		return dbmodel.Alert{}, fmt.Errorf("delete alert: %w", err)
	}
	return dbAlert, nil
}

func (c *client) getDBAlertAtom(id uint) (dbmodel.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return dbmodel.Alert{}, err
	}
	var dbAlert dbmodel.Alert
	err := c.dbAlertPreloaded().First(&dbAlert, id).Error
	if err != nil {
		return dbmodel.Alert{}, err
	}
	return dbAlert, nil
}

func (c *client) dbAlertPreloaded() *gorm.DB {
	return c.db.Model(&dbmodel.Alert{}).
		Preload(dbmodel.AlertFieldPlainMessage).
		Preload(dbmodel.AlertFieldAFK)
}
