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
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (c *client) MigrationStatus(ctx context.Context) (MigrationVersion, error) {
	if err := c.assertConnected(); err != nil {
		return MigrationUnknown, err
	}
	return c.withContext(ctx).migrationStatus()
}

func (c *client) migrationStatus() (MigrationVersion, error) {
	if c.prevMigChecked {
		return c.prevMigVersion, nil
	}
	status, err := getMigrationStatus(c.db)
	if err != nil {
		return MigrationUnknown, err
	}
	c.prevMigVersion = status
	c.prevMigChecked = true
	return status, nil
}

func getMigrationStatus(db *gorm.DB) (MigrationVersion, error) {
	var latest Migration
	silentDB := db.Session(&gorm.Session{Logger: logger.Discard})
	if err := silentDB.First(&latest).Error; err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrError &&
			strings.HasPrefix(sqliteErr.Error(), "no such table:") {
			return MigrationNeverApplied, nil
		}
		m := db.Migrator()
		if !m.HasTable(&Migration{}) {
			return MigrationNeverApplied, nil
		}
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
