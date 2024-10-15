package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type TeamRepository interface {
	CreateTeam(team *model.Team) (string, error)
	ReadTeams() (*[]model.Team, error)
	ReadTeamByID(teamId int64) (*model.Team, error)
	UpdateTeam(team *model.Team) error
	DeleteTeamByID(teamId int64) error
}

var _ TeamRepository = (*DBService)(nil)

func (s *DBService) CreateTeam(team *model.Team) (string, error) {
	var insertedID string
	query := "INSERT INTO public.team(name, password) VALUES($1, $2) RETURNING id"
	err := s.DB.QueryRow(query, team.Name, team.Password).Scan(&insertedID)
	if err != nil {
		return "", fmt.Errorf("failed team creation: %v", err)
	}
	return insertedID, nil
}

func (s *DBService) ReadTeams() (*[]model.Team, error) {
	var teams []model.Team
	query := "SELECT * FROM public.team"
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed getting teams: %v", err)
	}
	for rows.Next() {
		var r model.Team
		err := rows.Scan(&r.Id, &r.Name, &r.Password)
		teams = append(teams, r)
		if err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
	}
	return &teams, nil
}

func (s *DBService) ReadTeamByID(teamId int64) (*model.Team, error) {
	team := new(model.Team)
	query := "SELECT * FROM public.team WHERE id=$1"
	row := s.DB.QueryRow(query, teamId)
	err := row.Scan(&team.Id, &team.Name, &team.Password)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DBService) UpdateTeam(team *model.Team) error {
	row := s.DB.QueryRow("UPDATE public.team SET name=$1, password=$2 WHERE id=$3 RETURNING name,password", team.Name, team.Password, team.Id)
	err := row.Scan(&team.Name, &team.Password)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteTeamByID(teamId int64) error {
	_, err := s.DB.Exec("DELETE FROM public.team WHERE id=$1", teamId)
	if err != nil {
		return err
	}
	return nil
}
