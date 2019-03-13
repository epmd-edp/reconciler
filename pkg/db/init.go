package db

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitConnection() (*sql.DB, error) {
	host, present := os.LookupEnv("DB_HOST")
	if !present {
		return nil, errors.New("env variable DB_HOST is not present")
	}
	port, present := os.LookupEnv("DB_PORT")
	if !present {
		return nil, errors.New("env variable DB_PORT is not present")
	}
	name, present := os.LookupEnv("DB_NAME")
	if !present {
		return nil, errors.New("env variable DB_NAME is not present")
	}
	user, present := os.LookupEnv("DB_USER")
	if !present {
		return nil, errors.New("env variable DB_USER is not present")
	}
	pass, present := os.LookupEnv("DB_PASS")
	if !present {
		return nil, errors.New("env variable DB_PASS is not present")
	}
	ssl, present := os.LookupEnv("DB_SSL_MODE")
	if !present {
		return nil, errors.New("env variable DB_SSL_MODE is not present")
	}

	conn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v",
		host, port, name, user, pass, ssl)

	db, err := sql.Open("postgres", conn)

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	return db, nil
}
