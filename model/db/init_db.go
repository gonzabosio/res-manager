package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var ErrReadQuery = errors.New("failed query reading")

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		return nil, fmt.Errorf("connection to database failed: %v", err)
	}
	wg := sync.WaitGroup{}
	chanErr := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		q, err := os.ReadFile("model/db/team.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q))
		q2, err := os.ReadFile("model/db/project.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q2))
		q3, err := os.ReadFile("model/db/section.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q3))
		q4, err := os.ReadFile("model/db/resource.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q4))
		q5, err := os.ReadFile("model/db/participant.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q5))
		q6, err := os.ReadFile("model/db/user.sql")
		if err != nil {
			chanErr <- fmt.Errorf("%v: %v", ErrReadQuery, err)
			return
		}
		db.Exec(string(q6))
		chanErr <- nil
	}()
	wg.Wait()
	close(chanErr)
	if err := <-chanErr; err != nil {
		return nil, err
	}
	log.Println("Database connected successfully")
	return db, nil
}
