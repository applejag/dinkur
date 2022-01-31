// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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
	if err := c.withContext(ctx).migrate(); err != nil {
		return fmt.Errorf("migration: %w", err)
	}
	return nil
}

func (c *client) migrate() error {
	return c.transaction(func(tx *client) error {
		return tx.migrateNoTran()
	})
}

func (c *client) migrateNoTran() error {
	oldVersion, err := c.migrationStatus()
	if err != nil {
		return fmt.Errorf("check migration status: %w", err)
	}
	log.Debug().WithStringer("status", oldVersion).Message("Migration status checked.")
	if oldVersion == MigrationUpToDate {
		return nil
	}
	var start time.Time
	if oldVersion != MigrationNeverApplied {
		start = time.Now()
		log.Info().
			WithInt("old", int(oldVersion)).
			WithInt("new", int(LatestMigrationVersion)).
			Message("The database is outdated. Migrating...")
	}
	if oldVersion != MigrationNeverApplied && oldVersion < 5 {
		if err := c.db.Migrator().RenameTable("tasks", "entries"); err != nil {
			return err
		}
	}
	tables := []any{
		Migration{},
		Entry{},
		Alert{},
		AlertAFK{},
		AlertPlainMessage{},
		// Note: Do not add EntryFTS5 to auto migration! It is created separately
		// through manual SQL queries down below.
	}
	for _, tbl := range tables {
		log.Debug().WithStringf("type", "%T", tbl).Message("Applying auto migrations.")
		if err := c.db.AutoMigrate(tbl); err != nil {
			return err
		}
	}
	if oldVersion < 6 && c.db.Migrator().HasTable("tasks_idx") {
		if err := c.db.Migrator().DropTable("tasks_idx"); err != nil {
			return err
		}
	}
	log.Debug().Message("Done with auto migrations.")
	if oldVersion < 4 || !c.db.Migrator().HasTable("entries_idx") {
		// Creates FTS5 (Sqlite free-text search) virtual table
		// and triggers to keep it up-to-date.
		// Lastly it feeds it data from existing entries table in case of old data.
		err = c.db.Exec(`
CREATE VIRTUAL TABLE entries_idx USING fts5(name, content='entries',
	tokenize="porter trigram"
);
CREATE TRIGGER entries_idx_insert AFTER INSERT ON entries BEGIN
	INSERT INTO entries_idx(rowid, name) VALUES (new.id, new.name);
END;
CREATE TRIGGER entries_idx_delete AFTER DELETE ON entries BEGIN
	INSERT INTO entries_idx(entries_idx, rowid, name) VALUES ('delete', old.id, old.name);
END;
CREATE TRIGGER entries_idx_update AFTER UPDATE ON entries BEGIN
	INSERT INTO entries_idx(entries_idx, rowid, name) VALUES ('delete', old.id, old.name);
	INSERT INTO entries_idx(rowid, name) VALUES (new.id, new.name);
END;
INSERT INTO entries_idx (rowid, name) SELECT id, name FROM entries;
`).Error
		if err != nil {
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
	if oldVersion != MigrationNeverApplied {
		dur := time.Since(start)
		log.Info().WithDuration("duration", dur).Message("Database migration complete.")
	}
	return nil
}
