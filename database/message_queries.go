package database

import (
	"database/sql"
	"log"
)

func CreateMessageTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS messages (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			title TEXT,
    			message TEXT,
    			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				likes INTEGER DEFUALT 0,
    			user_id INTEGER,
    			parent_id INTEGER DEFAULT NULL,
    			topic_id INTEGER NOT NULL,
    			FOREIGN KEY (user_id) REFERENCES users(id),
    			FOREIGN KEY (parent_id) REFERENCES messages(id),
    			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error creating message table: ", err)
	}
}
