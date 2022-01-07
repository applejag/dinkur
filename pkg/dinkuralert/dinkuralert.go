// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package dinkuralert

import (
	"sync"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Store is a Dinkur alert store, which keeps track of alert IDs and provides
// an observable channel for alert updates.
type Store struct {
	observer
	lastID      uint
	lastIDMutex sync.Mutex

	afkActiveTask    *dinkur.Task
	afkAlert         *dinkur.Alert
	formerlyAFKAlert *dinkur.Alert
}

// Alerts returns a slice of all alerts.
func (s *Store) Alerts() []dinkur.Alert {
	var alerts []dinkur.Alert
	if s.afkAlert != nil {
		alerts = append(alerts, *s.afkAlert)
	}
	if s.formerlyAFKAlert != nil {
		alerts = append(alerts, *s.formerlyAFKAlert)
	}
	return alerts
}

// Delete removes an alert by ID.
func (s *Store) Delete(id uint) error {
	if s.afkAlert != nil && s.afkAlert.ID == id {
		s.PubAlertWait(AlertEvent{
			Alert: *s.afkAlert,
			Event: dinkur.EventDeleted,
		})
		s.afkAlert = nil
	} else if s.formerlyAFKAlert != nil && s.formerlyAFKAlert.ID == id {
		s.PubAlertWait(AlertEvent{
			Alert: *s.formerlyAFKAlert,
			Event: dinkur.EventDeleted,
		})
		s.formerlyAFKAlert = nil
	}
	return nil
}

// SetAFK marks the user as AFK and creates the AFK alert if it doesn't exist,
// as well as deleting the formerly-AFK alert if it exists.
func (s *Store) SetAFK(activeTask dinkur.Task) {
	if s.formerlyAFKAlert != nil {
		s.Delete(s.formerlyAFKAlert.ID)
	}
	if s.afkAlert != nil {
		return
	}
	now := time.Now()
	alert := dinkur.Alert{
		CommonFields: dinkur.CommonFields{
			ID:        s.nextID(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Type: dinkur.AlertAFK{
			ActiveTask: activeTask,
		},
	}
	s.afkActiveTask = &activeTask
	s.afkAlert = &alert
	s.PubAlertWait(AlertEvent{
		Alert: alert,
		Event: dinkur.EventCreated,
	})
}

// SetFormerlyAFK marks the user as formerly-AFK and creates the formerly-AFK
// alert if it doesn't exist, as well as deleting the AFK alert if it exists.
func (s *Store) SetFormerlyAFK(afkSince time.Time) {
	if s.afkAlert != nil {
		s.Delete(s.afkAlert.ID)
	}
	if s.afkActiveTask == nil || s.formerlyAFKAlert != nil {
		return
	}
	now := time.Now()
	alert := dinkur.Alert{
		CommonFields: dinkur.CommonFields{
			ID:        s.nextID(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Type: dinkur.AlertFormerlyAFK{
			AFKSince:   afkSince,
			ActiveTask: *s.afkActiveTask,
		},
	}
	s.formerlyAFKAlert = &alert
	s.PubAlertWait(AlertEvent{
		Alert: alert,
		Event: dinkur.EventCreated,
	})
}

func (s *Store) nextID() uint {
	s.lastIDMutex.Lock()
	s.lastID++
	id := s.lastID
	s.lastIDMutex.Unlock()
	return id
}
