package repository

import (
	"database/sql"
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
