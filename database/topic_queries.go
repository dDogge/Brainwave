package database

import (
	"database/sql"
	"errors"
	"fmt"
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

func AddTopic(db *sql.DB, title, username string) error {
	var creator_id int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creator_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("User not found")
		} else {
			log.Fatal(err)
		}
		log.Printf("Error fetching creator_id: %v", err)
		return fmt.Errorf("Could not fetch creator_id: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO topics (title, creator_id) VALUES (?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return fmt.Errorf("Could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(title, creator_id)
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.New("Topic tilte already exists")
		}
		log.Printf("Error executing statement: %v", err)
		return fmt.Errorf("Could not execute statement: %w", err)
	}

	log.Println("Topic added successfully:", title)
	return nil
}
