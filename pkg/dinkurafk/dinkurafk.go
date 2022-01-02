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

// Package dinkurafk contains code to detect if the user has gone AFK or
// returned from AFK.
package dinkurafk

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Errors specific to AFK-detectors.
var (
	ErrObserverIsNil = errors.New("observer is nil")
)

// Detector is an AFK-detector.
type Detector interface {
	// Start makes this detector start listening for OS-specific events to then
	// trigger AFK-started or AFK-stopped events.
	//
	// You need to stop the detector to remove any dangling Goroutines.
	StartDetecting() error
	// StopDetecting makes this detector stop listening for OS-specific events by
	// cleaning up its Goroutines and hooks.
	StopDetecting() error

	ObserverStarted
	ObserverStopped
}

// New creates a new AFK-detector.
func New() Detector {
	return &detector{
		ObserverStarted: NewObserverStarted(),
		ObserverStopped: NewObserverStopped(),
	}
}

type detector struct {
	ObserverStarted
	ObserverStopped

	isAFKMutex sync.RWMutex
	afkSince   *time.Time
}

func (d *detector) isAFK() bool {
	d.isAFKMutex.RLock()
	isAFK := d.afkSince != nil
	d.isAFKMutex.RUnlock()
	return isAFK
}

func (d *detector) setIsAFK(isAFK bool) (time.Time, bool) {
	if d.isAFK() == isAFK {
		// no update needed
		return time.Time{}, false
	}
	d.isAFKMutex.Lock()
	defer d.isAFKMutex.Unlock()
	if isAFK {
		now := time.Now()
		d.afkSince = &now
		return now, true
	}
	prevTimeSince := *d.afkSince
	d.afkSince = nil
	return prevTimeSince, true
}

func (d *detector) markAsAFK() {
	if _, changed := d.setIsAFK(true); !changed {
		return
	}
	d.ObserverStarted.PubStartedWait(Started{})
}

func (d *detector) markAsNoLongerAFK() {
	t, changed := d.setIsAFK(false)
	if !changed {
		return
	}
	d.ObserverStopped.PubStoppedWait(Stopped{
		AFKSince: t,
	})
}

func (d *detector) StartDetecting() error {
	return errors.New("not implemented")
}

func (d *detector) StopDetecting() error {
	unsubStartErr := d.UnsubAllStarted()
	unsubStopErr := d.UnsubAllStopped()
	if unsubStartErr != nil && unsubStopErr != nil {
		return fmt.Errorf("unsub all afk-start and stop subs: %w; %v", unsubStartErr, unsubStopErr)
	} else if unsubStartErr != nil {
		return fmt.Errorf("unsub all afk-start subs: %w", unsubStartErr)
	} else if unsubStopErr != nil {
		return fmt.Errorf("unsub all afk-stop subs: %w", unsubStopErr)
	}
	return nil
}
