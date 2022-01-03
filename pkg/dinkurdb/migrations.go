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
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (c *client) MigrationStatus(ctx context.Context) (MigrationVersion, error) {
	if err := c.assertConnected(); err != nil {
		return MigrationUnknown, err
	}
	return c.withContext(ctx).migrationStatus()
}

func (c *client) migrationStatus() (MigrationVersion, error) {
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

func getMigrationStatus(db *gorm.DB) (MigrationVersion, error) {
	m := db.Migrator()
	if !m.HasTable(&Migration{}) {
		return MigrationNeverApplied, nil
	}
	var latest Migration
	if err := db.First(&latest).Error; err != nil {
		return MigrationUnknown, nilNotFoundError(err)
	}
	v := MigrationVersion(latest.Version)
	return v, nil
}

func (c *client) Migrate(ctx context.Context) error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	return c.withContext(ctx).migrate()
}

func (c *client) migrate() error {
	return c.transaction(func(tx *client) error {
		return tx.migrateNoTran()
	})
}

func (c *client) migrateNoTran() error {
	migVersion, err := c.migrationStatus()
	if err != nil {
		return fmt.Errorf("check migration status: %w", err)
	}
	log.Debug().WithStringer("status", migVersion).Message("Migration status checked.")
	if migVersion == MigrationUpToDate {
		return nil
	}
	var start time.Time
	if migVersion != MigrationNeverApplied {
		start = time.Now()
		log.Info().
			WithInt("old", int(migVersion)).
			WithInt("new", int(LatestMigrationVersion)).
			Message("The database is outdated. Migrating...")
	}
	tables := []interface{}{
		&Migration{},
		&Task{},
	}
	for _, tbl := range tables {
		if err := c.db.AutoMigrate(tbl); err != nil {
			return err
		}
	}
	var migration Migration
	if err := c.db.FirstOrCreate(&migration).Error; err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	migration.Version = LatestMigrationVersion
	if err := c.db.Save(&migration).Error; err != nil {
		return err
	}
	if migVersion != MigrationNeverApplied {
		dur := time.Since(start)
		log.Info().WithDuration("duration", dur).Message("Database migration complete.")
	}
	return nil
}
