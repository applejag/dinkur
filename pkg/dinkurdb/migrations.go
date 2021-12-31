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

	"github.com/dinkur/dinkur/pkg/dinkur"
	"gorm.io/gorm"
)

// LatestMigrationVersion is an integer revision identifier for what migration
// was last applied to the database. This is stored in the database to quickly
// figure out if new migrations needs to be applied.
const LatestMigrationVersion = 2

func (c *client) MigrationStatus(ctx context.Context) (dinkur.MigrationStatus, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.MigrationUnknown, err
	}
	return c.withContext(ctx).migrationStatus()
}

func (c *client) migrationStatus() (dinkur.MigrationStatus, error) {
	if c.prevMigStatus != dinkur.MigrationUnknown {
		return c.prevMigStatus, nil
	}
	status, err := getMigrationStatus(c.db)
	if err != nil {
		return dinkur.MigrationUnknown, err
	}
	c.prevMigStatus = status
	return status, nil
}

func getMigrationStatus(db *gorm.DB) (dinkur.MigrationStatus, error) {
	m := db.Migrator()
	if !m.HasTable(&Migration{}) {
		return dinkur.MigrationNeverApplied, nil
	}
	var latest Migration
	if err := db.First(&latest).Error; err != nil {
		return dinkur.MigrationUnknown, nilNotFoundError(err)
	}
	if latest.Version < LatestMigrationVersion {
		return dinkur.MigrationOutdated, nil
	}
	return dinkur.MigrationUpToDate, nil
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
	status, err := c.migrationStatus()
	if err != nil {
		return fmt.Errorf("check migration status: %w", err)
	}
	log.Debug().WithStringer("status", status).Message("Migration status checked.")
	if status == dinkur.MigrationUpToDate {
		return nil
	}
	log.Info().Message("The database is outdated. Migrating...")
	start := time.Now()
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
	dur := time.Since(start)
	log.Info().WithDuration("duration", dur).Message("Database migration complete.")
	return nil
}
