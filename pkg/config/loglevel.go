// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
//
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

package config

import (
	"encoding"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"github.com/spf13/pflag"
)

type LogLevel logger.Level

func _() {
	// Ensure the type implements the interfaces
	l := LogLevel(logger.LevelDebug)
	var _ pflag.Value = &l
	var _ encoding.TextUnmarshaler = &l
	var _ jsonSchemaInterface = l
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	return l.Set(string(text))
}

func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l LogLevel) String() string {
	switch logger.Level(l) {
	case logger.LevelDebug:
		return "debug"
	case logger.LevelInfo:
		return "info"
	case logger.LevelWarn:
		return "warn"
	case logger.LevelError:
		return "error"
	case logger.LevelPanic:
		return "panic"
	case logger.LevelSilence:
		return "silence"
	default:
		return fmt.Sprintf("%[1]T(%[1]d)", logger.Level(l))
	}
}

func (l *LogLevel) Set(value string) error {
	lvl, err := logger.ParseLevel(value)
	if err != nil {
		return err
	}
	*l = LogLevel(lvl)
	return nil
}

func (l *LogLevel) Type() string {
	return "level"
}

// JSONSchema returns the JSON schema struct for this struct.
func (LogLevel) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging level",
		Enum: []any{
			LogLevel(logger.LevelDebug),
			LogLevel(logger.LevelInfo),
			LogLevel(logger.LevelWarn),
			LogLevel(logger.LevelError),
			LogLevel(logger.LevelPanic),
		},
	}
}
