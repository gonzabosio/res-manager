package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type ProjectRepository interface {
	CreateProject(proj *model.Project) (int64, error)
	ReadProject(projectId int64) (*model.Project, error)
	ReadProjectByTeamID(teamId int64) (*[]model.Project, error)
	UpdateProject(project *model.Project) error
	DeleteProjectByID(projectId int64) error
}

var _ ProjectRepository = (*DBService)(nil)

func (s *DBService) CreateProject(proj *model.Project) (int64, error) {
	var insertedID int64
	query := "INSERT INTO public.project(name, details, team_id) VALUES($1, $2, $3) RETURNING id"
	err := s.DB.QueryRow(query, proj.Name, proj.Details, proj.TeamId).Scan(&insertedID)
	if err != nil {
		return 0, err
	}
	return insertedID, nil
}

func (s *DBService) ReadProject(projectId int64) (*model.Project, error) {
	var project model.Project
	query := "SELECT * FROM public.project WHERE id=$1"
	row := s.DB.QueryRow(query, projectId)
	row.Scan(&project.Name, &project.Details, &project.TeamId)
	return &project, nil
}

func (s *DBService) ReadProjectByTeamID(teamId int64) (*[]model.Project, error) {
	var projs []model.Project
	query := "SELECT * FROM public.project WHERE team_id=$1"
	rows, err := s.DB.Query(query, teamId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r model.Project
		err := rows.Scan(&r.Id, &r.Name, &r.Details, &r.TeamId)
		projs = append(projs, r)
		if err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
	}
	return &projs, nil
}

func (s *DBService) UpdateProject(project *model.Project) error {
	row := s.DB.QueryRow("UPDATE public.project SET name=$1, details=$2 WHERE id=$3 RETURNING name,details", project.Name, project.Details, project.Id)
	err := row.Scan(&project.Name, &project.Details)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteProjectByID(projectId int64) error {
	_, err := s.DB.Exec("DELETE FROM public.project WHERE id=$1", projectId)
	if err != nil {
		return err
	}
	return nil
}
