package database

import (
	"database/sql"
	"log"
)

func CreateTopicTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS topics (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			title TEXT UNIQUE NOT NULL,
    			messages INTEGER DEFAULT 0,
    			creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    			creator_id INTEGER NOT NULL,
    			FOREIGN KEY (creator_id) REFERENCES users(id)
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error creating topic table: ", err)
	}
}
