package main

import (
	"log"

	"github.com/dinkur/dinkur/pkg/dinkurdb"
)

func main() {
	c := dinkurdb.NewClient()
	if err := c.Connect("dinkurdb.db"); err != nil {
		log.Fatalln("Error connecting to DB:", err)
	}
	if err := c.Ping(); err != nil {
		log.Fatalln("Error pinging DB:", err)
	}
	log.Println("Ping OK.")

	migration, err := c.MigrationStatus()
	if err != nil {
		log.Fatalln("Error checking migration status:", err)
	}
	log.Println("Migration status:", migration)

	if err := c.Migrate(); err != nil {
		log.Fatalln("Error migrating:", err)
	}

	migration, err = c.MigrationStatus()
	if err != nil {
		log.Fatalln("Error checking migration status:", err)
	}
	log.Println("Migration status:", migration)

	task, err := c.ActiveTask()
	if err != nil {
		log.Fatalln("Error getting active task:", err)
	}
	log.Printf("Active task: %+v", task)
}
