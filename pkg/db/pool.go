package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
)

var Instance *sql.DB

func init() {
	host := getEnvOrFatal("DB_HOST")
	port := getEnvOrFatal("DB_PORT")
	name := getEnvOrFatal("DB_NAME")
	user := getEnvOrFatal("DB_USER")
	pass := getEnvOrFatal("DB_PASS")
	ssl := getEnvOrFatal("DB_SSL_MODE")

	conn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v application_name=Reconciler",
		host, port, name, user, pass, ssl)

	db, err := sql.Open("postgres", conn)

	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	maxOpen := getIntEnvOrDefault("DB_MAX_OPEN_CONN", "5")
	maxIdle := getIntEnvOrDefault("DB_MAX_IDLE_CONN", "5")

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	Instance = db
}

func getEnvOrFatal(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("env variable %v is missing", key)
	}
	return value
}

func getIntEnvOrDefault(key, defaultValue string) int {
	strVal := getEnvOrDefault(key, defaultValue)
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		log.Fatalf("cannot convert env value %v to int", strVal)
	}
	return intVal
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
