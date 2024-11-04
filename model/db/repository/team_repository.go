package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gonzabosio/gobo-patcher"
	"github.com/gonzabosio/res-manager/model"
	"golang.org/x/crypto/bcrypt"
)

type TeamRepository interface {
	CreateTeam(*model.Team) (int64, error)
	ReadTeams() (*[]model.Team, error)
	ReadTeamByName(*model.Team) error
	UpdateTeam(*model.PatchTeam) error
	DeleteTeamByID(int64) error
}

var _ TeamRepository = (*DBService)(nil)

func (s *DBService) CreateTeam(team *model.Team) (int64, error) {
	query := "SELECT id,password FROM public.team WHERE name=$1"
	var pw string
	err := s.DB.QueryRow(query, team.Name).Scan(&team.Id, &pw)
	if err != nil {
		if err == sql.ErrNoRows {
			var insertedID int64
			insert := "INSERT INTO public.team(name, password) VALUES($1, $2) RETURNING id"
			err = s.DB.QueryRow(insert, team.Name, team.Password).Scan(&insertedID)
			if err != nil {
				return 0, fmt.Errorf("failed team creation: %v", err)
			}
			return insertedID, nil
		} else {
			return 0, err
		}
	}
	return 0, fmt.Errorf("team already exists")
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
		if err := rows.Scan(&r.Id, &r.Name, &r.Password); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		teams = append(teams, r)
	}
	return &teams, nil
}

func (s *DBService) ReadTeamByName(team *model.Team) error {
	query := "SELECT id,password FROM public.team WHERE name=$1"
	var pw string
	err := s.DB.QueryRow(query, team.Name).Scan(&team.Id, &pw)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("team not found")
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte(team.Password)); err != nil {
		return fmt.Errorf("invalid team data: %v", err)
	}
	team.Password = pw
	return nil
}

func (s *DBService) UpdateTeam(team *model.PatchTeam) error {
	origTeam := model.Team{}
	s.DB.QueryRow("SELECT * FROM public.team WHERE id=$1", team.Id).
		Scan(&origTeam.Id, &origTeam.Name, &origTeam.Password)
	orB, err := json.Marshal(origTeam)
	if err != nil {
		return fmt.Errorf("could not parse project struct to json: %v", err)
	}
	newB, err := json.Marshal(team)
	if err != nil {
		return fmt.Errorf("could not parse team struct to json: %v", err)
	}
	log.Println(string(orB), string(newB))
	query, err := gobo.PatchWithQuery(orB, newB, "public.team", "id", true, nil)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("%v RETURNING name, password", query)
	if err = s.DB.QueryRow(q).Scan(&team.Name, &team.Password); err != nil {
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
