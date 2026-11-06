package main

import (
	"log"

	"tracker/internal/db"
)

func main() {
	if err := db.ResetMigrations("migrations", "tracker.db"); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations complete")
}
