package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func CreateUserTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS users (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			username TEXT UNIQUE NOT NULL,
    			password TEXT NOT NULL,
				email TEXT NOT NULL,
    			topics_opened INTEGER DEFAULT 0,
    			messages_sent INTEGER DEFAULT 0,
    			creation_date DATETIME DEFAULT CURRENT_TIMESTAMP
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("error creating user table: ", err)
	}
}

func AddUser(db *sql.DB, username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return fmt.Errorf("could not hash password: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, email, hashedPassword)
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.New("username or email already exists")
		}
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	log.Println("user added successfully:", username)
	return nil
}

func CheckPassword(db *sql.DB, username, password string) (bool, error) {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

func ChangePassword(db *sql.DB, username, currentPassword, newPassword string) error {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword))
	if err != nil {
		return errors.New("incorrect current password")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", hashedNewPassword, username)
	if err != nil {
		return err
	}

	return nil
}

func ChangeEmail(db *sql.DB, username, newEmail string) error {
	var existingUser string
	err := db.QueryRow("SELECT username FROM users WHERE email = ?", newEmail).Scan(&existingUser)
	if err != sql.ErrNoRows {
		return errors.New("email is already in use")
	}

	_, err = db.Exec("UPDATE users SET email = ? WHERE username = ?", newEmail, username)
	if err != nil {
		return err
	}

	return nil
}

func ChangeUsername(db *sql.DB, username, newUsername string) error {
	var existingUser string
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&existingUser)
	if err != sql.ErrNoRows {
		return errors.New("username is already in use")
	}

	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !validUsername.MatchString(newUsername) {
		return errors.New("username does not meet requirement")
	}

	_, err = db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, username)
	if err != nil {
		return err
	}

	return nil
}

func RemoveUser(db *sql.DB, username string) error {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user not found: %s", username)
			return fmt.Errorf("user not found: %s", username)
		}
		log.Printf("error fetching user ID: %v", err)
		return fmt.Errorf("could not fetch user ID: %w", err)
	}

	_, err = db.Exec("UPDATE topics SET user_id = NULL WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("error setting user_id to NULL in topics: %v", err)
		return fmt.Errorf("could not set user_id to NULL in topics: %w", err)
	}

	_, err = db.Exec("UPDATE messages SET user_id = NULL WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("error setting user_id to NULL in messages: %v", err)
		return fmt.Errorf("could not set user_id to NULL in messages: %w", err)
	}

	stmt, err := db.Prepare("DELETE FROM users WHERE username = ?")
	if err != nil {
		log.Printf("error preparing statement: %v", err)
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(username)
	if err != nil {
		log.Printf("error executing statement: %v", err)
		return fmt.Errorf("could not execute statement: %w", err)
	}

	log.Println("user removed successfully:", username)
	return nil
}

func isUniqueConstraintError(err error) bool {
	return err != nil && (err.Error() == "UNIQUE constraint failed: users.username" || err.Error() == "UNIQUE constraint failed: users.email")
}
