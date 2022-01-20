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

package obs

import (
	"errors"
	"sync"
	"time"
)

// Errors specific for the listener and subscriptions.
var (
	ErrAlreadyUnsubscribed       = errors.New("already unsubscribed")
	ErrSubscriptionNotInitalized = errors.New("subscription is not initialized")
)

func New[T any]() Observer[T] {
	return &observer[T]{}
}

type Observer[T any] interface {
	Pub(T)
	Sub() <-chan T
	Unsub(<-chan T) error
	UnsubAll() error
}

type observer[T any] struct {
	subs      []chan T
	mutex     sync.RWMutex
	OnFailPub func(T)
}

func (o *observer[T]) Pub(ev T) {
	o.mutex.RLock()
	for _, sub := range o.subs {
		go func(ev T, sub chan T) {
			select {
			case sub <- ev:
			case <-time.After(10 * time.Second):
				o.OnFailPub(ev)
			}
		}(ev, sub)
	}
	o.mutex.RUnlock()
}

// Sub subscribes to events in a newly created channel.
func (o *observer[T]) Sub() <-chan T {
	o.mutex.Lock()
	sub := make(chan T)
	o.subs = append(o.subs, sub)
	o.mutex.Unlock()
	return sub
}

// Unsub unsubscribes a previously subscribed channel.
func (o *observer[T]) Unsub(sub <-chan T) error {
	if sub == nil {
		return ErrSubscriptionNotInitalized
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	idx := o.subIndex(sub)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	close(o.subs[idx])
	o.subs = append(o.subs[:idx], o.subs[idx+1:]...)
	return nil
}

// UnsubAll unsubscribes all subscription channels, rendering them all useless.
func (o *observer[T]) UnsubAll() error {
	o.mutex.Lock()
	for _, ch := range o.subs {
		close(ch)
	}
	o.subs = nil
	o.mutex.Unlock()
	return nil
}

func (o *observer[T]) subIndex(sub <-chan T) int {
	for i, ch := range o.subs {
		if ch == sub {
			return i
		}
	}
	return -1
}
