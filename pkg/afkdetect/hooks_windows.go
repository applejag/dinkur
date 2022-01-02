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

package afkdetect

// #include "hooks_windows.h"
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var singletonWindowsHooks = &windowsHooks{}

func init() {
	detectorHooks = append(detectorHooks, singletonWindowsHooks)
}

type windowsHooks struct {
	detector *detector
	detMutex sync.RWMutex
}

func (h *windowsHooks) Register(d *detector) error {
	if d == nil {
		return nil
	}
	h.detMutex.Lock()
	defer h.detMutex.Unlock()
	if h.detector != nil {
		return errors.New("only 1 windows hooks can be registered at a time")
	}
	log.Debug().Message("Registering Windows hooks WH_KEYBOARD_LL & WH_MOUSE_LL.")
	h.detector = d
	if err := convRegisterCodeToErr(int32(C.RegisterHooks())); err != nil {
		return err
	}
	return nil
}

func (h *windowsHooks) Unregister(d *detector) error {
	h.detMutex.Lock()
	defer h.detMutex.Unlock()
	if d == nil || h.detector == nil {
		return nil
	}
	if h.detector != d {
		return errors.New("not the same detector")
	}
	if err := convUnregisterCodeToErr(int32(C.UnregisterHooks())); err != nil {
		return err
	}
	log.Debug().Message("Unregistering Windows hooks WH_KEYBOARD_LL & WH_MOUSE_LL.")
	h.detector = nil
	return nil
}

func convRegisterCodeToErr(code int32) error {
	switch code {
	case 1:
		return errors.New("already registered hooks")
	default:
		return convSysErrCode(code)
	}
}

func convUnregisterCodeToErr(code int32) error {
	switch code {
	case 1:
		return errors.New("already unregistered hooks")
	default:
		return convSysErrCode(code)
	}
}

func (h *windowsHooks) Tick() error {
	if bool(C.GetWorkstationLocked()) {
		h.detector.markAsAFK()
		return nil
	}
	if err := convSysErrCode(int32(C.GetThreadStatus())); err != nil {
		return err
	}
	sinceAFKMs := C.GetTickMs() - C.GetLastEventTickMs()
	sinceAFK := (time.Duration(sinceAFKMs) * time.Millisecond).Truncate(time.Second)
	if sinceAFK > afkThresholdDur {
		h.detector.markAsAFK()
	} else {
		h.detector.markAsNoLongerAFK()
	}
	return nil
}

//export goTriggerTick
func goTriggerTick() {
	singletonWindowsHooks.Tick()
}

func convSysErrCode(code int32) error {
	switch code {
	case 0, 259: // 259 = STILL_ACTIVE (thread status)
		return nil
	case 1404:
		return errors.New("1404 (0x57C): invalid hook handle")
	case 1426:
		return errors.New("1426 (0x592): invalid hook procedure type")
	case 1427:
		return errors.New("1427 (0x593): invalid hook procedure")
	case 1428:
		return errors.New("1428 (0x594): cannot set nonlocal hook without a module handle")
	case 1429:
		return errors.New("1429 (0x595): this hook procedure can only be set globally")
	case 1430:
		return errors.New("1430 (0x596): the journal hook procedure is already installed")
	case 1431:
		return errors.New("1431 (0x597): the hook procedure is not installed")
	case 1458:
		return errors.New("1458 (0x5B2): hook type not allowed")
	default:
		return fmt.Errorf("unknown Windows system error: %d (0x%[1]X)", code)
	}
}
