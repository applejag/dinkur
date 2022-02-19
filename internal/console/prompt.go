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

package console

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/mattn/go-isatty"
)

func checkIfNonInteractiveTTY() bool {
	return os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))
}

func convPromptErr(err error) error {
	if err == io.EOF {
		return fmt.Errorf("%w (maybe you are piping STDIN?)", err)
	}
	return err
}

// PromptEntryRemoval asks the user for confirmation about removing a entry.
// Will return an io.EOF error if the current TTY is not an interactive session.
func PromptEntryRemoval(entry dinkur.Entry) error {
	var sb strings.Builder
	promptWarnIconColor.Fprint(&sb, promptWarnIconText)
	sb.WriteByte(' ')
	sb.WriteString("Warning: You are about to permanently remove entry ")
	writeEntryID(&sb, entry.ID)
	sb.WriteByte(' ')
	writeEntryName(&sb, entry.Name)
	sb.WriteByte('.')
	fmt.Fprintln(stderr, sb.String())
	var ok bool
	prompt := &survey.Confirm{
		Message: "Are you sure?",
	}
	if err := survey.AskOne(prompt, &ok); err != nil {
		return convPromptErr(err)
	}
	if !ok {
		fmt.Println("Aborted by user.")
		os.Exit(1)
	}
	return nil
}

// AFKResolution states what should be changed as decided from the human's AFK
// resolution.
type AFKResolution struct {
	Edit     *dinkur.EditEntry
	NewEntry *dinkur.NewEntry
}

// PromptAFKResolution asks the user for how to resolve an AFK status.
func PromptAFKResolution(activeEntry dinkur.Entry, afkSince, backSince time.Time) (AFKResolution, error) {
	var sb strings.Builder
	now := time.Now()

	promptWarnIconColor.Fprint(&sb, promptWarnIconText)
	sb.WriteString(" Note: You were away since ")
	writeEntryTimeSpanNow(&sb, afkSince, nil)
	sb.WriteByte(' ')
	writeEntryDurationWithDelim(&sb, now.Sub(afkSince))
	sb.WriteByte('\n')
	promptWarnIconColor.Fprint(&sb, promptWarnIconText)
	sb.WriteString(" while having an active entry ")
	writeEntryID(&sb, activeEntry.ID)
	sb.WriteByte(' ')
	writeEntryName(&sb, activeEntry.Name)
	sb.WriteByte(' ')
	writeEntryTimeSpanActiveDuration(&sb, activeEntry.Start, activeEntry.End, activeEntry.Elapsed())
	sb.WriteString("\n\n")

	if checkIfNonInteractiveTTY() {
		promptWarnIconColor.Fprint(&sb, promptWarnIconText)
		sb.WriteString(" The terminal seems to be non-interactive. Skipping prompt.\n")
		promptWarnIconColor.Fprint(&sb, promptWarnIconText)
		sb.WriteString(` Assuming option "1. Leave the active entry as-is and continue with the invoked command."`)
		fmt.Fprintln(stderr, sb.String())
		entryEditNoneColor.Fprintln(stdout, entryEditPrefix, entryEditNoChange)
		return AFKResolution{}, nil
	}

	sb.WriteString("How do you want to save this away time?\n")

	sb.WriteString("  1. Leave the active entry as-is and continue with the invoked command.\n")

	sb.WriteString("  2. Discard the away time I was away, changing active entry to ")
	writeEntryTimeSpanNowDuration(&sb, activeEntry.Start, &afkSince, afkSince.Sub(activeEntry.Start))
	sb.WriteString(".\n")

	sb.WriteString("  3. Save the away time as a new entry ")
	writeEntryTimeSpanNowDuration(&sb, afkSince, nil, now.Sub(afkSince))
	sb.WriteString("  (naming it in a later prompt).\n")

	sb.WriteByte(' ')
	promptCtrlCHelpColor.Fprint(&sb, "(press Ctrl+C to abort)")
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	fmt.Fprint(stderr, sb.String())

	prompt := &survey.Input{
		Message: "Select option [1-3]:",
	}
	answerInt, err := promptIntRange(prompt, 1, 3)
	if err != nil {
		return AFKResolution{}, err
	}

	switch answerInt {
	case 1:
		// Leave the active entry as-is.
		entryEditNoneColor.Fprintln(stdout, entryEditPrefix, entryEditNoChange)
		return AFKResolution{}, nil

	case 2:
		// Discard the time
		fmt.Fprintln(stderr, "Discarding the away time from the currently active entry.")
		return AFKResolution{
			Edit: &dinkur.EditEntry{
				IDOrZero: activeEntry.ID,
				End:      &afkSince,
			},
		}, nil

	case 3:
		// Save the time as a new entry
		return promptAFKSaveAsNewEntry(activeEntry, afkSince)

	default:
		return AFKResolution{}, errors.New("no answer chosen")
	}
}

func promptAFKSaveAsNewEntry(activeEntry dinkur.Entry, afkSince time.Time) (AFKResolution, error) {
	name, err := promptNonEmptyString(&survey.Input{
		Message: "Enter name of new entry:",
	})
	if err != nil {
		return AFKResolution{}, err
	}
	var sb strings.Builder
	sb.WriteString("Saving the away time as a new entry with name ")
	writeEntryName(&sb, name)
	sb.WriteString(".\n")
	fmt.Fprint(stderr, sb.String())
	return AFKResolution{
		Edit: &dinkur.EditEntry{
			IDOrZero: activeEntry.ID,
			End:      &afkSince,
		},
		NewEntry: &dinkur.NewEntry{
			Name:               name,
			Start:              &afkSince,
			StartAfterIDOrZero: activeEntry.ID,
		},
	}, nil
}

func promptNonEmptyString(prompt survey.Prompt) (string, error) {
	for {
		var answer string
		if err := survey.AskOne(prompt, &answer); err != nil {
			return "", convPromptErr(err)
		}

		if answer == "" {
			promptErrorColor.Fprintf(stderr, "Please enter a value.\n\n")
			continue
		}

		return answer, nil
	}
}

func promptIntRange(prompt survey.Prompt, lower, upper int) (int, error) {
	for {
		answer, err := promptNonEmptyString(prompt)
		if err != nil {
			return 0, err
		}

		answerInt, err := strconv.Atoi(answer)
		if err != nil {
			promptErrorColor.Fprintf(stderr, "Invalid answer: %v\n\n", err)
			continue
		}

		if answerInt < lower || answerInt > upper {
			promptErrorColor.Fprintf(stderr, "Please enter a value in the range %d-%d.\n\n", lower, upper)
			continue
		}

		return answerInt, nil
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
