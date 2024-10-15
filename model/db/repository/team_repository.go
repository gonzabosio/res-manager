package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type TeamRepository interface {
	CreateTeam(team *model.Team) (string, error)
}

var _ TeamRepository = (*DBService)(nil)

func (s *DBService) CreateTeam(team *model.Team) (string, error) {
	var insertedID string
	query := `INSERT INTO public.team(name, password) VALUES($1, $2) RETURNING id`
	err := s.DB.QueryRow(query, team.Name, team.Password).Scan(&insertedID)
	if err != nil {
		return "", fmt.Errorf("failed team creation: %v", err)
	}
	return insertedID, nil
}
