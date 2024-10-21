package repository

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type DBService struct {
	DB *sql.DB
}

type RepositoryService struct {
	TeamRepository
	ProjectRepository
	SectionRepository
	ResourceRepository
	ParticipantRepository
}

func comparePassword(hashedPassword, password []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashedPassword, password); err != nil {
		return err
	}
	return nil
}
