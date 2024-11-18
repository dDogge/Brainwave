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
				upvotes INTEGER DEFAULT 0,
    			creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    			creator_id INTEGER,
    			FOREIGN KEY (creator_id) REFERENCES users(id)
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("error creating topic table: ", err)
	}
}

func AddTopic(db *sql.DB, title, username string) error {
	var creatorID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		log.Printf("error fetching creator_id: %v", err)
		return fmt.Errorf("could not fetch creator_id: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO topics (title, creator_id) VALUES (?, ?)")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(title, creatorID)
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.New("topic title already exists")
		}
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	_, err = db.Exec("UPDATE users SET topics_opened = topics_opened + 1 WHERE id = ?", creatorID)
	if err != nil {
		log.Printf("error incrementing topics_opened for user ID %d: %v", creatorID, err)
		return fmt.Errorf("could not increment topics_opened: %w", err)
	}

	log.Println("topic added successfully:", title)
	return nil
}

func RemoveTopic(db *sql.DB, title string) error {
	stmt, err := db.Prepare("DELETE FROM topics WHERE title = ?")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(title)
	if err != nil {
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("error retrieving rows affected: %v", err)
		return fmt.Errorf("could not retrieve rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("no topic found with title: %s", title)
		return fmt.Errorf("no topic found with title: %s", title)
	}

	log.Println("topic removed successfully:", title)
	return nil
}

func UpVoteTopic(db *sql.DB, title, username string) error {
	stmt, err := db.Prepare("UPDATE topics SET upvotes = upvotes + 1 WHERE title = ?")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(title)
	if err != nil {
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	log.Println("upvote added successfully for topic:", title)
	return nil
}

func DownVoteTopic(db *sql.DB, title, username string) error {
	stmt, err := db.Prepare("UPDATE topics SET upvotes = upvotes - 1 WHERE title = ?")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(title)
	if err != nil {
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	log.Println("downvote added successfully for topic:", title)
	return nil
}

func GetAllTopics(db *sql.DB) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, title, messages, upvotes, creation_date, creator_id FROM topics")
	if err != nil {
		log.Printf("error fetching topics: %v", err)
		return nil, fmt.Errorf("could not fetch topics: %w", err)
	}
	defer rows.Close()

	var topics []map[string]interface{}
	for rows.Next() {
		var id, messages, upvotes, creatorID sql.NullInt64
		var title, creationDate string

		if err := rows.Scan(&id, &title, &messages, &upvotes, &creationDate, &creatorID); err != nil {
			log.Printf("error scanning topic row: %v", err)
			return nil, fmt.Errorf("could not scan topic row: %w", err)
		}

		topic := map[string]interface{}{
			"id":            id.Int64,
			"title":         title,
			"messages":      messages.Int64,
			"upvotes":       upvotes.Int64,
			"creation_date": creationDate,
			"creator_id":    creatorID.Int64,
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

func GetTopicByTitle(db *sql.DB, title string) (map[string]interface{}, error) {
	var id, messages, upvotes, creatorID sql.NullInt64
	var creationDate string

	err := db.QueryRow("SELECT id, messages, upvotes, creation_date, creator_id FROM topics WHERE title = ?", title).
		Scan(&id, &messages, &upvotes, &creationDate, &creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("topic with title '%s' not found", title)
		}
		log.Printf("error fetching topic by title '%s': %v", title, err)
		return nil, fmt.Errorf("could not fetch topic: %w", err)
	}

	topic := map[string]interface{}{
		"id":            id.Int64,
		"title":         title,
		"messages":      messages.Int64,
		"upvotes":       upvotes.Int64,
		"creation_date": creationDate,
		"creator_id":    creatorID.Int64,
	}

	return topic, nil
}
