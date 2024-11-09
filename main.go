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

	database.AddUser(db, "dDogge", "blabla1@email.com", "oijeoidfwiuo")
	database.AddUser(db, "Imbus", "blabla2@email.com", "odegrfwegeriuo")
	database.ChangePassword(db, "Imbus", "odegrfwegeriuo", "yoyoyo")
	database.ChangeUsername(db, "Imbus", "Imbus64")
	database.ChangeEmail(db, "Imbus64", "dwasdw@email.com")
}
