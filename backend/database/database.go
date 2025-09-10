package database

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Global variable for the database connection
var DB *sql.DB

// Connect to the PostgreSQL database
func ConnectDB() {
	connectionString := "user=queue_app dbname=queue_app password=queue_app sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Check if the connection is actually working
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to the database!")
}