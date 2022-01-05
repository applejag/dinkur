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

package console

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

// PromptTaskRemoval asks the user for confirmation about removing a task.
// Will return an io.EOF error if the current TTY is not an interactive session.
func PromptTaskRemoval(task dinkur.Task) (bool, error) {
	var sb strings.Builder
	promptWarnIconColor.Fprint(&sb, promptWarnIconText)
	sb.WriteByte(' ')
	sb.WriteString("Warning: You are about to permanently remove task ")
	writeTaskID(&sb, task.ID)
	sb.WriteByte(' ')
	writeTaskName(&sb, task.Name)
	sb.WriteByte('.')
	fmt.Fprintln(stderr, sb.String())
	var ok bool
	prompt := &survey.Confirm{
		Message: "Are you sure?",
	}
	if err := survey.AskOne(prompt, &ok); err != nil {
		if err == io.EOF {
			return false, fmt.Errorf("%w (maybe you are piping STDIN?)", err)
		}
		return false, err
	}
	return ok, nil
}
