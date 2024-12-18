package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"regexp"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func CreateUserTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS users (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			username TEXT UNIQUE NOT NULL,
    			password TEXT NOT NULL,
				email TEXT NOT NULL,
				reset_code TEXT DEFAULT NULL,
    			topics_opened INTEGER DEFAULT 0,
    			messages_sent INTEGER DEFAULT 0,
    			creation_date DATETIME DEFAULT CURRENT_TIMESTAMP
			);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("error creating user table: ", err)
		return err
	}
	return nil
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
			log.Printf("unique constraint violation for username or email: %v", err)
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
			log.Printf("user not found: %s", username)
			return false, nil
		}
		log.Printf("error fetching password for username %s: %v", username, err)
		return false, fmt.Errorf("could not fetch password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("incorrect password for user %s: %v", username, err)
		return false, nil
	}

	return true, nil
}

func ChangePassword(db *sql.DB, username, currentPassword, newPassword string) error {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user not found: %s", username)
			return errors.New("user not found")
		}
		log.Printf("error fetching password for user %s: %v", username, err)
		return fmt.Errorf("could not fetch password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword))
	if err != nil {
		log.Printf("incorrect current password for user %s: %v", username, err)
		return errors.New("incorrect current password")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing new password: %v", err)
		return fmt.Errorf("could not hash new password: %w", err)
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", hashedNewPassword, username)
	if err != nil {
		log.Printf("error updating password for user %s: %v", username, err)
		return fmt.Errorf("could not update password: %w", err)
	}

	log.Println("password changed successfully for user:", username)
	return nil
}

func ChangeEmail(db *sql.DB, username, newEmail string) error {
	var existingUser string
	err := db.QueryRow("SELECT username FROM users WHERE email = ?", newEmail).Scan(&existingUser)
	if err == nil {
		return errors.New("email is already in use")
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("could not check for existing email: %w", err)
	}

	if err == nil {
		log.Printf("email already in use: %s", newEmail)
		return errors.New("email is already in use")
	}

	_, err = db.Exec("UPDATE users SET email = ? WHERE username = ?", newEmail, username)
	if err != nil {
		log.Printf("error updating email for user %s: %v", username, err)
		return fmt.Errorf("could not update email: %w", err)
	}

	log.Println("email updated successfully for user:", username)
	return nil
}

func ChangeUsername(db *sql.DB, username, newUsername string) error {
	var existingUser string
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&existingUser)
	if err != sql.ErrNoRows {
		log.Printf("error checking for existing username %s: %v", newUsername, err)
		return fmt.Errorf("could not check for existing username: %w", err)
	}

	if err == nil {
		log.Printf("username already in use: %s", newUsername)
		return errors.New("username is already in use")
	}

	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !validUsername.MatchString(newUsername) {
		log.Printf("invalid username format: %s", newUsername)
		return errors.New("username does not meet requirements")
	}

	_, err = db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, username)
	if err != nil {
		log.Printf("error updating username for user %s: %v", username, err)
		return fmt.Errorf("could not update username: %w", err)
	}

	log.Println("username updated successfully from", username, "to", newUsername)
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

	_, err = db.Exec("UPDATE topics SET creator_id = NULL WHERE creator_id = ?", userID)
	if err != nil {
		log.Printf("error setting creator_id to NULL in topics: %v", err)
		return fmt.Errorf("could not set creator_id to NULL in topics: %w", err)
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

func GeneratePasswordResetCode(db *sql.DB, email string) (string, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("email not found")
		}
		log.Printf("error fetching user ID for email %s: %v", email, err)
		return "", fmt.Errorf("could not fetch user ID: %w", err)
	}

	var code int = rand.IntN(900000) + 100000
	var resetCode string = strconv.Itoa(code)
	_, err = db.Exec("UPDATE users SET reset_code = ? WHERE id = ?", resetCode, userID)
	if err != nil {
		log.Printf("error creating password reset code for user ID %d: %v", userID, err)
		return "", fmt.Errorf("could not create password reset code: %w", err)
	}

	return resetCode, nil
}

func ResetPassword(db *sql.DB, email, resetCode, newPassword string) error {
	var userID int
	err := db.QueryRow("SELECT id from users WHERE reset_code = ?", resetCode).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid reset code")
		}
		log.Printf("error fetching user ID for reset code %s: %v", resetCode, err)
		return fmt.Errorf("could not fetch user ID: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing new password: %v", err)
		return fmt.Errorf("could not hash new password: %w", err)
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		log.Printf("error updating password for user ID %d: %v", userID, err)
		return fmt.Errorf("could not update password: %w", err)
	}

	_, err = db.Exec("UPDATE users SET reset_code = NULL WHERE id = ?", userID)
	if err != nil {
		log.Printf("error deleting password reset code for user ID %d: %v", userID, err)
		return fmt.Errorf("could not delete password reset code: %w", err)
	}

	return nil
}

func GetAllUsers(db *sql.DB) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, username, email, topics_opened, messages_sent, creation_date FROM users")
	if err != nil {
		log.Printf("error fetching all users: %v", err)
		return nil, fmt.Errorf("could not fetch users: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, topicsOpened, messagesSent sql.NullInt64
		var username, email, creationDate string

		if err := rows.Scan(&id, &username, &email, &topicsOpened, &messagesSent, &creationDate); err != nil {
			log.Printf("error scanning user row: %v", err)
			return nil, fmt.Errorf("could not scan user row: %w", err)
		}

		user := map[string]interface{}{
			"id":            id.Int64,
			"username":      username,
			"email":         email,
			"topics_opened": topicsOpened.Int64,
			"messages_sent": messagesSent.Int64,
			"creation_date": creationDate,
		}
		users = append(users, user)
	}

	return users, nil
}
