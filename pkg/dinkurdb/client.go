package dinkurdb

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Client interface {
	Connect(filename string) error
	Ping() error
	Close() error

	Migrate() error
	MigrationStatus() (MigrationStatus, error)

	StartTask(task NewTask) (StartedTask, error)
	ActiveTask() (*Task, error)
	StopActiveTask() (bool, error)
}

func NewClient() Client {
	return &client{}
}

type client struct {
	db *gorm.DB
}

func (c *client) Connect(sqliteDSN string) (err error) {
	if c.db != nil {
		return ErrAlreadyConnected
	}
	c.db, err = gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{
		Logger: logger.Discard,
		//Logger: logger.Default.LogMode(logger.Info),
	})

	return
}

func (c *client) Ping() error {
	if c.db == nil {
		return ErrNotConnected
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}

func (c *client) Close() error {
	if c.db == nil {
		return ErrNotConnected
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	if err := sql.Close(); err != nil {
		return err
	}
	c.db = nil
	return nil
}
