// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

package dinkurdb

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const LatestMigrationVersion = 1

type MigrationStatus byte

const (
	MigrationUnknown MigrationStatus = iota
	MigrationNeverApplied
	MigrationOutdated
	MigrationUpToDate
)

func (s MigrationStatus) String() string {
	switch s {
	case MigrationUnknown:
		return "unknown"
	case MigrationNeverApplied:
		return "never applied"
	case MigrationOutdated:
		return "outdated"
	case MigrationUpToDate:
		return "up to date"
	default:
		return fmt.Sprintf("%T(%d)", s, s)
	}
}

func (c *client) MigrationStatus() (MigrationStatus, error) {
	if c.db == nil {
		return MigrationUnknown, ErrNotConnected
	}
	if c.prevMigStatus != MigrationUnknown {
		return c.prevMigStatus, nil
	}
	status, err := getMigrationStatus(c.db)
	if err != nil {
		return MigrationUnknown, err
	}
	c.prevMigStatus = status
	return status, nil
}

func getMigrationStatus(db *gorm.DB) (MigrationStatus, error) {
	m := db.Migrator()
	if !m.HasTable(&Migration{}) {
		return MigrationNeverApplied, nil
	}
	var latest Migration
	if err := db.First(&latest).Error; err != nil {
		return MigrationUnknown, nilNotFoundError(err)
	}
	if latest.Version < LatestMigrationVersion {
		return MigrationOutdated, nil
	}
	return MigrationUpToDate, nil
}

func (c *client) Migrate() error {
	if c.db == nil {
		return ErrNotConnected
	}
	return c.transaction(func(tx *client) error {
		status, err := tx.MigrationStatus()
		if err != nil {
			return fmt.Errorf("check migration status: %w", err)
		}
		if status == MigrationUpToDate {
			return nil
		}
		tables := []interface{}{
			&Migration{},
			&Task{},
		}
		for _, tbl := range tables {
			if err := tx.db.AutoMigrate(tbl); err != nil {
				return err
			}
		}
		var migration Migration
		if err := tx.db.FirstOrCreate(&migration).Error; err != nil &&
			!errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		migration.Version = LatestMigrationVersion
		return tx.db.Save(&migration).Error
	})
}
