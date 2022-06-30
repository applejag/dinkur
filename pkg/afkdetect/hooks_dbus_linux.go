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

package afkdetect

import (
	"errors"
	"time"

	"github.com/godbus/dbus/v5"
)

func init() {
	detectorHooks = append(detectorHooks, dbusHookRegisterer{})
}

type dbusHookRegisterer struct {
}

func (h dbusHookRegisterer) Register(d *detector) (detectorHook, error) {
	if d == nil {
		return nil, nil
	}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Debug().WithError(err).Message("Failed to connect to session dbus.")
		return nil, nil // swallow error, in case of GNU/Linux distros w/o dbus
	}
	log.Debug().Message("Registering dbus connection for org.gnome.Mutter, org.gnome.ScreenSaver & org.gnome.Shell.")
	// https://unix.stackexchange.com/a/492328
	// https://gitlab.gnome.org/GNOME/mutter/-/blob/41.2/src/org.gnome.Mutter.IdleMonitor.xml#L14-16
	idleMon := conn.Object("org.gnome.Mutter.IdleMonitor", "/org/gnome/Mutter/IdleMonitor/Core")
	hook := &dbusHook{
		d:       d,
		conn:    conn,
		idleMon: idleMon,
	}
	// https://people.gnome.org/~mccann/gnome-screensaver/docs/gnome-screensaver.html#gs-signals
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.gnome.ScreenSaver"),
	); err != nil {
		return nil, err
	}

	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.gnome.Shell.Introspect"),
		dbus.WithMatchMember("RunningApplicationsChanged"),
	); err != nil {
		return nil, err
	}

	go hook.handleDbusSignal()

	return hook, nil
}

func (h *dbusHook) handleDbusSignal() {
	log.Debug().Message("Listen for dbus signals...")
	ch := make(chan *dbus.Signal, 10)
	h.conn.Signal(ch)
	for signal := range ch {
		switch signal.Name {
		case "org.gnome.ScreenSaver.ActiveChanged":
			if len(signal.Body) != 1 {
				continue
			}
			activeChanged, ok := signal.Body[0].(bool)
			if !ok {
				continue
			}
			if activeChanged {
				h.d.markAsAFK()
			} else {
				h.d.markAsNoLongerAFK()
			}
		case "org.gnome.ScreenSaver.WakeUpScreen",
			"org.gnome.Shell.Introspect.RunningApplicationsChanged":
			h.d.markAsNoLongerAFK()
		default:
			log.Debug().WithString("name", signal.Name).Message("Unknown dbus signal.")
		}
	}
}

type dbusHook struct {
	d       *detector
	conn    *dbus.Conn
	idleMon dbus.BusObject
}

func (h *dbusHook) Unregister() error {
	log.Debug().Message("Unregistering dbus connection.")
	return h.conn.Close()
}

func (h *dbusHook) Tick() error {
	if h.idleMon == nil {
		return nil
	}
	var idleDurMs uint64
	if err := h.idleMon.Call("org.gnome.Mutter.IdleMonitor.GetIdletime", 0).Store(&idleDurMs); err != nil {
		var dbusErr dbus.Error
		if errors.As(err, &dbusErr) && dbusErr.Name == "org.freedesktop.DBus.Error.ServiceUnknown" {
			log.Debug().WithError(err).
				Message("Detected 'unknown service' error. Disabling org.gnome.Mutter integration.")
			h.idleMon = nil
		} else {
			return err
		}
	}
	idleDur := time.Duration(idleDurMs) * time.Millisecond
	if idleDur > afkThresholdDur {
		h.d.markAsAFK()
	} else {
		h.d.markAsNoLongerAFK()
	}
	return nil
}
