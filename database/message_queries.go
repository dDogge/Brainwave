package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func CreateMessageTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS messages (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
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

func AddMessage(db *sql.DB, topic, message, username string) error {
	var creatorID int
	var topicID int

	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("User not found")
		}
		log.Printf("Error fetching creator_id: %v", err)
		return fmt.Errorf("could not fetch creator_id: %w", err)
	}

	err = db.QueryRow("SELECT id FROM topics WHERE title = ?", topic).Scan(&topicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Topic not found")
		}
		log.Printf("Error fetching creator_id: %v", err)
		return fmt.Errorf("could not fetch creator_id: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO topics (message, user_id, topic_id) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return fmt.Errorf("Could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(message, creatorID, topicID)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return fmt.Errorf("Could not execute statement: %w", err)
	}

	_, err = db.Exec("UPDATE users SET messages_sent = messages_sent + 1 WHERE id = ?", creatorID)
	if err != nil {
		log.Printf("Error incrementing messages_sent for user ID %d: %v", creatorID, err)
		return fmt.Errorf("Could not increment messages_sent: %w", err)
	}

	_, err = db.Exec("UPDATE topics SET messages = messages + 1 WHERE id = ?", topicID)
	if err != nil {
		log.Printf("Error incrementing messages for topic ID %d: %v", topicID, err)
		return fmt.Errorf("Could not increment messages: %w", err)
	}

	log.Println("Message added successfully:", message)
	return nil
}
