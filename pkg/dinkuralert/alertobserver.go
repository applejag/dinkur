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
	"errors"
	"sync"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Errors specific for the listener and subscriptions.
var (
	ErrAlreadyUnsubscribed       = errors.New("already unsubscribed")
	ErrSubscriptionNotInitalized = errors.New("subscription is not initialized")
)

// AlertEvent is an alert and its event type.
type AlertEvent struct {
	Alert dinkur.Alert
	Event dinkur.EventType
}

// NewObserver returns a new Dinkur alerts observer.
func NewObserver() Observer {
	return &observer{}
}

// Observer lets you publish and subscribe Dinkur alerts.
type Observer interface {
	// PubAlertWait publishes an alert and waits until all
	// subscriptions has received their events.
	PubAlertWait(AlertEvent)
	SubAlerts() <-chan AlertEvent
	UnsubAlerts(<-chan AlertEvent) error
	UnsubAllAlerts() error
}

type observer struct {
	subs  []chan AlertEvent
	mutex sync.RWMutex
}

func (o *observer) PubAlertWait(s AlertEvent) {
	var wg sync.WaitGroup
	o.mutex.RLock()
	wg.Add(len(o.subs))
	for _, sub := range o.subs {
		go func(s AlertEvent, sub chan AlertEvent, wg *sync.WaitGroup) {
			sub <- s
			wg.Done()
		}(s, sub, &wg)
	}
	o.mutex.RUnlock()
	wg.Wait()
}

func (o *observer) SubAlerts() <-chan AlertEvent {
	o.mutex.Lock()
	sub := make(chan AlertEvent)
	o.subs = append(o.subs, sub)
	o.mutex.Unlock()
	return sub
}

func (o *observer) UnsubAlerts(sub <-chan AlertEvent) error {
	if sub == nil {
		return ErrSubscriptionNotInitalized
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	idx := o.subIndex(sub)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	o.subs = append(o.subs[idx:], o.subs[idx+1:]...)
	return nil
}

func (o *observer) UnsubAllAlerts() error {
	o.mutex.Lock()
	o.subs = nil
	o.mutex.Unlock()
	return nil
}

func (o *observer) subIndex(sub <-chan AlertEvent) int {
	for i, ch := range o.subs {
		if ch == sub {
			return i
		}
	}
	return -1
}
