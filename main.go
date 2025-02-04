package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/dDogge/Brainwave/database"
	_ "modernc.org/sqlite"
)

// This type of embed assumes that the frontend is built and will hard fail otherwise
//
//go:embed frontend/dist
var frontend embed.FS

func main() {
	reactFS, err := fs.Sub(frontend, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	http.Handle("/", http.FileServer(http.FS(reactFS)))

	// var err error
	db, err := sql.Open("sqlite", "./brainwave_db.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Error enabling foreign key support:", err)
	}

	database.CreateUserTable(db)
	database.CreateMessageTable(db)
	database.CreateTopicTable(db)

	port := ":8080"
	log.Printf("Serving on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
