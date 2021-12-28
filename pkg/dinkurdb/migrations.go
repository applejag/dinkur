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
	m := c.db.Migrator()
	if !m.HasTable(&Migration{}) {
		return MigrationNeverApplied, nil
	}
	var latest Migration
	if err := c.db.First(&latest).Error; err != nil {
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
	status, err := c.MigrationStatus()
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
	return c.db.Save(&migration).Error
}
