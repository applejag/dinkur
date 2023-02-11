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
	"github.com/spf13/pflag"
)

type LogColor string

const (
	LogColorAuto   LogColor = "auto"
	LogColorNever  LogColor = "never"
	LogColorAlways LogColor = "always"
)

func _() {
	// Ensure the type implements the interfaces
	f := LogColorNever
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
	var _ jsonSchemaInterface = f
}

func (f LogColor) String() string {
	return string(f)
}

func (f *LogColor) Set(value string) error {
	switch LogColor(value) {
	case LogColorAuto:
		*f = LogColorAuto
	case LogColorNever, "no", "off", "false":
		*f = LogColorNever
	case LogColorAlways, "yes", "on", "true":
		*f = LogColorAlways
	default:
		return fmt.Errorf("unknown log format: %q, must be one of: auto, never, always", value)
	}
	return nil
}

func (f *LogColor) Type() string {
	return "format"
}

func (f *LogColor) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

// JSONSchema returns the JSON schema struct for this struct.
func (LogColor) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging coloring",
		Enum: []any{
			LogColorAuto,
			LogColorNever,
			LogColorAlways,
		},
		Default: LogColorAuto,
	}
}
