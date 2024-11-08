package main

import (
	"database/sql"
	"log"

	"github.com/dDogge/Brainwave/database"
	_ "modernc.org/sqlite"
)

func main() {
	var err error
	db, err := sql.Open("sqlite", "./brainwave_db.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	database.CreateUserTable(db)
	database.CreateMessageTable(db)
	database.CreateTopicTable(db)
}
