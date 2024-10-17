package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var ErrReadQuery = errors.New("failed query reading")

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		return nil, fmt.Errorf("connection to database failed: %v", err)
	}
	q, err := os.ReadFile("model/db/team.sql")
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrReadQuery, err)
	}
	db.Exec(string(q))
	q2, err := os.ReadFile("model/db/project.sql")
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrReadQuery, err)
	}
	db.Exec(string(q2))
	log.Println("Database connected successfully")
	return db, nil
}
