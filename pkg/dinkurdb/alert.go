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
	"gopkg.in/typ.v2"
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
	switch alert := newAlert.Alert.(type) {
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
	case nil:
		return dbmodel.Alert{}, errors.New("alert type was nil")
	default:
		return dbmodel.Alert{}, fmt.Errorf("unsupported alert type: %v", newAlert.Alert.Type())
	}
	if err := c.db.Create(&dbAlert).Error; err != nil {
		return dbmodel.Alert{}, err
	}
	return dbAlert, nil
}

type newOrUpdatedDBAlert struct {
	before *dbmodel.Alert // will be nil if alert was created
	after  dbmodel.Alert
}

func (c *client) CreateOrUpdateAlertByType(ctx context.Context, newAlert dinkur.NewAlert) (dinkur.NewOrUpdatedAlert, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.NewOrUpdatedAlert{}, err
	}
	newOrUpdate, err := c.withContext(ctx).createOrUpdateDBAlertByType(newAlert)
	if err != nil {
		return dinkur.NewOrUpdatedAlert{}, err
	}
	alertBefore, err := fromdb.AlertPtr(newOrUpdate.before)
	alertAfter, err := fromdb.Alert(newOrUpdate.after)
	return dinkur.NewOrUpdatedAlert{
		Before: alertBefore,
		After:  alertAfter,
	}, nil
}

func (c *client) createOrUpdateDBAlertByType(newAlert dinkur.NewAlert) (newOrUpdatedDBAlert, error) {
	var update newOrUpdatedDBAlert
	err := c.transaction(func(tx *client) (tranErr error) {
		update, tranErr = tx.createOrUpdateDBAlertByTypeNoTran(newAlert)
		return
	})
	return update, err
}

func (c *client) createOrUpdateDBAlertByTypeNoTran(newAlert dinkur.NewAlert) (newOrUpdatedDBAlert, error) {
	dbAlert, err := c.getDBAlertByTypeNoTran(newAlert.Alert.Type())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newDBAlert, err := c.createDBAlertAtom(dinkur.NewAlert{
			Alert: newAlert.Alert,
		})
		if err != nil {
			return newOrUpdatedDBAlert{}, fmt.Errorf("create new alert as no alert of type was found: %w", err)
		}
		return newOrUpdatedDBAlert{
			before: nil,
			after:  newDBAlert,
		}, nil
	} else if err != nil {
		return newOrUpdatedDBAlert{}, fmt.Errorf("get alert by type for create-or-update: %w", err)
	}
	update, err := c.editDBAlertNoTran(dbAlert, dinkur.EditAlert{
		ID:    dbAlert.ID,
		Alert: newAlert.Alert,
	})
	if err != nil {
		return newOrUpdatedDBAlert{}, fmt.Errorf("update existing alert by type: %w", err)
	}
	return newOrUpdatedDBAlert{
		before: &update.before,
		after:  update.after,
	}, nil
}

func (c *client) GetAlertList(ctx context.Context) ([]dinkur.Alert, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbAlerts, err := c.withContext(ctx).listDBAlerts()
	if err != nil {
		return nil, err
	}
	alerts, err := fromdb.AlertSlice(dbAlerts)
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

func (c *client) listDBAlerts() ([]dbmodel.Alert, error) {
	var dbAlerts []dbmodel.Alert
	err := c.transaction(func(tx *client) (tranErr error) {
		dbAlerts, tranErr = tx.listDBAlertsNoTran()
		return
	})
	return dbAlerts, err
}

func (c *client) listDBAlertsNoTran() ([]dbmodel.Alert, error) {
	var dbAlerts []dbmodel.Alert
	if err := c.dbAlertPreloaded().Find(&dbAlerts).Error; err != nil {
		return nil, err
	}
	dbAFKAlertIDs := typ.Map(dbAlerts, func(a dbmodel.Alert) uint { return a.ID })
	var dbAFKAlerts []dbmodel.AlertAFK
	if err := c.db.Preload(dbmodel.AlertAFKFieldActiveEntry).
		Find(&dbAFKAlerts, dbAFKAlertIDs).
		Error; err != nil {
		return nil, err
	}
	for i, alert := range dbAlerts {
		for _, afkAlert := range dbAFKAlerts {
			if alert.AFK == nil && alert.AFK.ID != afkAlert.ID {
				continue
			}
			dbAlerts[i].AFK = &afkAlert
		}
	}
	return dbAlerts, nil
}

func (c *client) UpdateAlert(ctx context.Context, edit dinkur.EditAlert) (dinkur.UpdatedAlert, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedAlert{}, err
	}
	dbUpdated, err := c.withContext(ctx).editDBAlert(edit)
	if err != nil {
		return dinkur.UpdatedAlert{}, err
	}
	dbAlertBefore, err := fromdb.Alert(dbUpdated.before)
	if err != nil {
		return dinkur.UpdatedAlert{}, err
	}
	dbAlertAfter, err := fromdb.Alert(dbUpdated.after)
	if err != nil {
		return dinkur.UpdatedAlert{}, err
	}
	return dinkur.UpdatedAlert{
		Before: dbAlertBefore,
		After:  dbAlertAfter,
	}, nil
}

type updatedDBAlert struct {
	before dbmodel.Alert
	after  dbmodel.Alert
}

func (c *client) editDBAlert(edit dinkur.EditAlert) (updatedDBAlert, error) {
	var update updatedDBAlert
	err := c.transaction(func(tx *client) (tranErr error) {
		dbAlert, err := tx.getDBAlertAtom(edit.ID)
		if err != nil {
			return fmt.Errorf("get alert to edit: %w", err)
		}
		update, tranErr = tx.editDBAlertNoTran(dbAlert, edit)
		return
	})
	return update, err
}

func (c *client) editDBAlertNoTran(dbAlert dbmodel.Alert, edit dinkur.EditAlert) (updatedDBAlert, error) {
	var (
		dbAlertToSave any
		changed       bool
	)
	switch alert := edit.Alert.(type) {
	case dinkur.AlertPlainMessage:
		if dbAlert.PlainMessage == nil {
			return updatedDBAlert{}, errors.New("cannot change alert type of existing alert")
		}
		dbAlertToSave, changed = c.editDBAlertPlainMessageNoTran(*dbAlert.PlainMessage, alert)
	case dinkur.AlertAFK:
		if dbAlert.AFK == nil {
			return updatedDBAlert{}, errors.New("cannot change alert type of existing alert")
		}
		dbAlertToSave, changed = c.editDBAlertAFKNoTran(*dbAlert.AFK, alert)
	case nil:
		return updatedDBAlert{}, errors.New("cannot change alert type to nil")
	default:
		return updatedDBAlert{}, fmt.Errorf("unknown alert type: %T", edit.Alert)
	}
	if !changed {
		return updatedDBAlert{
			before: dbAlert,
			after:  dbAlert,
		}, nil
	}
	if err := c.db.Save(dbAlertToSave).Error; err != nil {
		return updatedDBAlert{}, fmt.Errorf("saving changes to alert: %w", err)
	}
	dbAlertAfter, err := c.getDBAlertAtom(edit.ID)
	if err != nil {
		return updatedDBAlert{}, fmt.Errorf("get alert after edit: %w", err)
	}
	return updatedDBAlert{
		before: dbAlert,
		after:  dbAlertAfter,
	}, nil
}

func (c *client) editDBAlertPlainMessageNoTran(before dbmodel.AlertPlainMessage, edit dinkur.AlertPlainMessage) (dbmodel.AlertPlainMessage, bool) {
	if before.Message == edit.Message {
		return before, false
	}
	before.Message = edit.Message
	return before, true
}

func (c *client) editDBAlertAFKNoTran(before dbmodel.AlertAFK, edit dinkur.AlertAFK) (dbmodel.AlertAFK, bool) {
	if before.AFKSince.UTC() == edit.AFKSince.UTC() &&
		before.ActiveEntryID == edit.ActiveEntry.ID &&
		(before.BackSince == nil) == (edit.BackSince == nil) &&
		(before.BackSince != nil || before.BackSince.UTC() == edit.BackSince.UTC()) {
		return before, false
	}
	before.AFKSince = edit.AFKSince
	before.ActiveEntryID = edit.ActiveEntry.ID
	before.BackSince = conv.TimePtrUTC(edit.BackSince)
	return before, true
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
	return fromdb.Alert(dbAlert)
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
