package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func Connect() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "gcit_booking")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v\nMake sure PostgreSQL is running and the DB '%s' exists.", err, dbname)
	}

	Db = db
	log.Println("Connected to PostgreSQL database:", dbname)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
